// Package hubcli implements `hub`, the command line client for the hub API.
//
// Every command except the handful that manage configuration is generated from
// the API catalog, so the CLI gains a command the moment an rpc is added to a
// .proto - there is no table of endpoints to keep in step. That also makes the
// surface predictable enough for an AI agent to use without being told: rpc
// ListUser on UserService is `hub user list-user`, its request fields are its
// flags, and `hub api list` enumerates the lot as JSON.
package hubcli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

// Options configures a CLI instance.
type Options struct {
	// Catalog is the API surface to build commands from. Defaults to the
	// globally registered descriptors.
	Catalog *apicatalog.Catalog
	// Version is reported by `hub version`.
	Version string
}

// CLI holds the state shared by every command: the flags parsed off the root,
// and the settings and client derived from them once flag parsing is done.
type CLI struct {
	catalog *apicatalog.Catalog
	version string

	profile    string
	flags      Settings
	outputFlag string
	dryRun     bool

	settings Settings
	format   Format
	client   *Client
}

// NewRootCommand builds the `hub` command tree.
func NewRootCommand(opts Options) *cobra.Command {
	catalog := opts.Catalog
	if catalog == nil {
		catalog = apicatalog.Default()
	}
	c := &CLI{catalog: catalog, version: opts.Version}

	root := &cobra.Command{
		Use:   "hub",
		Short: "Command line client for the hub API",
		Long: `Command line client for the hub API.

The command tree is generated from the protobuf definitions, so every rpc is
reachable as "hub <service> <method>" with one flag per request field. Run
"hub api list" for the machine-readable catalog and "hub api describe <rpc>"
for a single operation.

Configuration is layered: the profile in ~/.hub/config.yaml, then HUB_ENDPOINT,
HUB_TOKEN and friends, then the flags below.`,
		SilenceUsage: true,
		// Execute reports the error itself, with a single consistent prefix.
		SilenceErrors:     true,
		PersistentPreRunE: c.prepare,
	}

	flags := root.PersistentFlags()
	flags.StringVar(&c.profile, "profile", "", "configuration profile to use")
	flags.StringVar(&c.flags.Endpoint, "endpoint", "", "base URL of the hub API (env HUB_ENDPOINT)")
	flags.StringVar(&c.flags.Token, "token", "", "bearer token (env HUB_TOKEN)")
	flags.StringVarP(&c.outputFlag, "output", "o", string(FormatJSON), "output format: json, yaml or table")
	flags.BoolVar(&c.dryRun, "dry-run", false, "print the request that would be sent instead of sending it")

	root.AddCommand(
		c.newAPICommand(),
		c.newAuthCommand(),
		c.newConfigCommand(),
		c.newVersionCommand(),
	)
	root.AddCommand(c.newServiceCommands()...)
	return root
}

// Execute runs the CLI and returns the process exit code.
func Execute(opts Options) int {
	if err := NewRootCommand(opts).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}

// prepare resolves the settings once the root flags are parsed, so that every
// command below can rely on c.client and c.format.
func (c *CLI) prepare(cmd *cobra.Command, _ []string) error {
	format, err := ParseFormat(c.outputFlag)
	if err != nil {
		return err
	}
	c.format = format

	path, err := ConfigPath()
	if err != nil {
		return err
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		return err
	}

	c.settings = Resolve(cfg.Profile(c.profile), c.flags)
	c.client = NewClient(c.settings)
	cmd.SilenceUsage = true
	return nil
}

// newServiceCommands turns the catalog into one command group per service.
func (c *CLI) newServiceCommands() []*cobra.Command {
	groups := map[string]*cobra.Command{}
	var names []string

	for _, op := range c.catalog.Operations() {
		if op.HTTPMethod == "" {
			continue
		}
		name := GroupName(op)
		group, ok := groups[name]
		if !ok {
			group = &cobra.Command{
				Use:   name,
				Short: fmt.Sprintf("Operations of %s", op.Service),
			}
			groups[name] = group
			names = append(names, name)
		}
		group.AddCommand(c.newOperationCommand(op))
	}

	sort.Strings(names)
	commands := make([]*cobra.Command, 0, len(names))
	for _, name := range names {
		commands = append(commands, groups[name])
	}
	return commands
}

func (c *CLI) newOperationCommand(op apicatalog.Operation) *cobra.Command {
	cmd := &cobra.Command{
		Use:     CommandName(op),
		Aliases: CommandAliases(op),
		Short:   op.Summary,
		Long:    operationHelp(op),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			values, err := collectFieldValues(cmd, op)
			if err != nil {
				return err
			}
			return c.send(cmd, op, values)
		},
	}
	registerFieldFlags(cmd, op)
	return cmd
}

// send performs the call, or prints it when --dry-run is set.
func (c *CLI) send(cmd *cobra.Command, op apicatalog.Operation, values map[string]any) error {
	req, err := c.client.BuildRequest(op, values)
	if err != nil {
		return err
	}
	if c.dryRun {
		return Render(cmd.OutOrStdout(), c.format, req)
	}
	response, err := c.client.Do(cmd.Context(), req)
	if err != nil {
		return err
	}
	return Render(cmd.OutOrStdout(), c.format, response)
}

func operationHelp(op apicatalog.Operation) string {
	help := fmt.Sprintf("%s\n\nrpc:      %s.%s\nendpoint: %s %s\n",
		op.Summary, op.Service, op.Method, op.HTTPMethod, op.Path)
	if op.Public {
		return help + "rbac:     none, any authenticated caller may use this rpc\n"
	}
	return help + fmt.Sprintf("rbac:     %s on %s\n", op.Action, op.Resource)
}

func (c *CLI) newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			version := c.version
			if version == "" {
				version = "dev"
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
}
