package hubcli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

func (c *CLI) newAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Explore and call the API directly",
	}
	cmd.AddCommand(c.newAPIListCommand(), c.newAPIDescribeCommand(), c.newAPICallCommand(), c.newAPIDocsCommand())
	return cmd
}

// operationSummary is the shape `hub api list` and `hub api describe` emit. It
// is the CLI's contract with anything that reads it - an agent deciding which
// endpoint to call, most of all - so the field names are spelled out here
// rather than borrowed from the catalog's internal struct.
type operationSummary struct {
	Service    string       `json:"service" yaml:"service"`
	Method     string       `json:"method" yaml:"method"`
	Command    string       `json:"command" yaml:"command"`
	HTTPMethod string       `json:"httpMethod" yaml:"httpMethod"`
	Path       string       `json:"path" yaml:"path"`
	Summary    string       `json:"summary" yaml:"summary"`
	Public     bool         `json:"public" yaml:"public"`
	Resource   string       `json:"resource" yaml:"resource"`
	Action     string       `json:"action" yaml:"action"`
	Fields     []fieldModel `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type fieldModel struct {
	Name       string   `json:"name" yaml:"name"`
	Flag       string   `json:"flag" yaml:"flag"`
	In         string   `json:"in" yaml:"in"`
	Kind       string   `json:"kind" yaml:"kind"`
	Repeated   bool     `json:"repeated,omitempty" yaml:"repeated,omitempty"`
	EnumValues []string `json:"enumValues,omitempty" yaml:"enumValues,omitempty"`
}

func summarise(op apicatalog.Operation, withFields bool) operationSummary {
	s := operationSummary{
		Service:    op.Service,
		Method:     op.Method,
		Command:    fmt.Sprintf("hub %s %s", GroupName(op), CommandName(op)),
		HTTPMethod: op.HTTPMethod,
		Path:       op.Path,
		Summary:    op.Summary,
		Public:     op.Public,
		Resource:   op.Resource,
		Action:     op.Action,
	}
	if !withFields {
		return s
	}
	for _, field := range op.Fields {
		s.Fields = append(s.Fields, fieldModel{
			Name:       field.JSONName,
			Flag:       "--" + FlagName(field),
			In:         string(field.In),
			Kind:       field.Kind,
			Repeated:   field.Repeated,
			EnumValues: field.EnumValues,
		})
	}
	return s
}

func (c *CLI) newAPIListCommand() *cobra.Command {
	var service string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List every API operation",
		Long: `List every API operation with its REST mapping, RBAC rule and CLI command.

This is the catalog an agent should read first: it names the endpoint, the
permission it needs and the exact command that calls it.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var operations []operationSummary
			for _, op := range c.catalog.Operations() {
				if service != "" && op.Service != service && op.ServiceName != service {
					continue
				}
				operations = append(operations, summarise(op, false))
			}
			if operations == nil {
				return fmt.Errorf("no operations match service %q", service)
			}
			return Render(cmd.OutOrStdout(), c.format, map[string]any{"operations": operations})
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "only list operations of this service")
	return cmd
}

func (c *CLI) newAPIDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe <operation>",
		Short: "Describe one API operation and its request fields",
		Long: `Describe one API operation and its request fields.

The operation may be named as "ListUser", "UserService.ListUser",
"user.v1.UserService.ListUser" or "/user.v1.UserService/ListUser".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			op, err := c.catalog.Find(args[0])
			if err != nil {
				return err
			}
			return Render(cmd.OutOrStdout(), c.format, summarise(op, true))
		},
	}
}

func (c *CLI) newAPICallCommand() *cobra.Command {
	var (
		data  string
		query []string
	)
	cmd := &cobra.Command{
		Use:   "call <method> <path>",
		Short: "Call an arbitrary API path",
		Long: `Call an arbitrary API path.

The escape hatch for anything the generated commands do not cover. The path is
taken verbatim, so it must include the /api/v1 prefix.

  hub api call GET /api/v1/users --query limit=10 --query status=STATUS_ACTIVE
  hub api call POST /api/v1/groups --data '{"name":"platform"}'
  hub api call PUT /api/v1/users/u1 --data @user.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			if !isHTTPMethod(method) {
				return fmt.Errorf("unsupported HTTP method %q", method)
			}

			body, err := readData(data)
			if err != nil {
				return err
			}
			values, err := parseQuery(query)
			if err != nil {
				return err
			}

			req := Request{Method: method, URL: strings.TrimRight(c.settings.Endpoint, "/") + args[1], Body: body}
			if encoded := values.Encode(); encoded != "" {
				req.URL += "?" + encoded
			}
			if c.dryRun {
				return Render(cmd.OutOrStdout(), c.format, req)
			}
			if err := c.ensureFreshToken(cmd.Context()); err != nil {
				return err
			}
			response, err := c.client.Do(cmd.Context(), req)
			if err != nil {
				return err
			}
			return Render(cmd.OutOrStdout(), c.format, response)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body, or @file to read it from a file")
	cmd.Flags().StringArrayVar(&query, "query", nil, "query parameter as key=value, repeat for repeated fields")
	return cmd
}

func isHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func readData(data string) (json.RawMessage, error) {
	if data == "" {
		return nil, nil
	}
	raw := []byte(data)
	if path, ok := strings.CutPrefix(data, "@"); ok {
		contents, err := os.ReadFile(path) //nolint:gosec // the caller names the file
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		raw = contents
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("--data is not valid JSON")
	}
	return raw, nil
}

func parseQuery(pairs []string) (url.Values, error) {
	values := url.Values{}
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("--query %q is not in key=value form", pair)
		}
		values.Add(key, value)
	}
	return values, nil
}

func (c *CLI) newAPIDocsCommand() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Write the API reference the agent skill ships",
		Long: `Write the API reference the agent skill ships.

Regenerated by "make gen" so the reference an agent reads cannot drift from the
protos the server actually serves.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			reference := RenderReference(c.catalog)
			if out == "" {
				_, err := cmd.OutOrStdout().Write([]byte(reference))
				return err
			}
			if err := os.WriteFile(out, []byte(reference), 0o600); err != nil {
				return fmt.Errorf("write %s: %w", out, err)
			}
			_, err := fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", out)
			return err
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "file to write, defaults to stdout")
	return cmd
}
