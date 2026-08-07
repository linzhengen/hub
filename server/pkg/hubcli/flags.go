package hubcli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

// registerFieldFlags gives an operation one flag per request field, typed after
// the protobuf field so cobra rejects a bad value before a request is sent.
func registerFieldFlags(cmd *cobra.Command, op apicatalog.Operation) {
	flags := cmd.Flags()
	for _, field := range op.Fields {
		name, usage := FlagName(field), fieldUsage(field)
		switch {
		case field.Kind == "message", field.Kind == "map", field.Repeated && field.Kind != "string":
			flags.String(name, "", usage)
		case field.Repeated:
			flags.StringSlice(name, nil, usage)
		case field.Kind == "bool":
			flags.Bool(name, false, usage)
		case field.Kind == "double", field.Kind == "float":
			flags.Float64(name, 0, usage)
		case strings.HasPrefix(field.Kind, "uint"), strings.HasPrefix(field.Kind, "fixed"):
			flags.Uint64(name, 0, usage)
		case strings.HasPrefix(field.Kind, "int"), strings.HasPrefix(field.Kind, "sint"), strings.HasPrefix(field.Kind, "sfixed"):
			flags.Int64(name, 0, usage)
		default:
			flags.String(name, "", usage)
		}
	}
}

func fieldUsage(field apicatalog.Field) string {
	parts := []string{string(field.In)}
	if field.Repeated {
		parts = append(parts, "repeated")
	}
	switch {
	case len(field.EnumValues) > 0:
		parts = append(parts, "one of "+strings.Join(field.EnumValues, "|"))
	case field.Kind == "map", field.Kind == "message":
		parts = append(parts, "JSON "+field.Kind)
	default:
		parts = append(parts, field.Kind)
	}
	return strings.Join(parts, ", ")
}

// collectFieldValues reads the flags the caller actually set. Untouched flags
// are omitted rather than sent as zero values, so an update does not silently
// blank a field the caller never mentioned.
func collectFieldValues(cmd *cobra.Command, op apicatalog.Operation) (map[string]any, error) {
	values := map[string]any{}
	for _, field := range op.Fields {
		name := FlagName(field)
		if !cmd.Flags().Changed(name) {
			continue
		}
		value, err := fieldValue(cmd.Flags(), name, field)
		if err != nil {
			return nil, fmt.Errorf("--%s: %w", name, err)
		}
		values[field.Name] = value
	}
	return values, nil
}

func fieldValue(flags *pflag.FlagSet, name string, field apicatalog.Field) (any, error) {
	switch {
	case field.Kind == "message", field.Kind == "map", field.Repeated && field.Kind != "string":
		raw, err := flags.GetString(name)
		if err != nil {
			return nil, err
		}
		var decoded any
		if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
			return nil, fmt.Errorf("expected JSON: %w", err)
		}
		return decoded, nil
	case field.Repeated:
		return flags.GetStringSlice(name)
	case field.Kind == "bool":
		return flags.GetBool(name)
	case field.Kind == "double", field.Kind == "float":
		return flags.GetFloat64(name)
	case strings.HasPrefix(field.Kind, "uint"), strings.HasPrefix(field.Kind, "fixed"):
		return flags.GetUint64(name)
	case strings.HasPrefix(field.Kind, "int"), strings.HasPrefix(field.Kind, "sint"), strings.HasPrefix(field.Kind, "sfixed"):
		return flags.GetInt64(name)
	default:
		raw, err := flags.GetString(name)
		if err != nil {
			return nil, err
		}
		if len(field.EnumValues) > 0 && !contains(field.EnumValues, raw) {
			return nil, fmt.Errorf("%q is not one of %s", raw, strings.Join(field.EnumValues, ", "))
		}
		return raw, nil
	}
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
