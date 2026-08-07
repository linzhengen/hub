package hubcli

import (
	"strings"
	"unicode"

	"github.com/linzhengen/hub/v1/server/pkg/apicatalog"
)

// crudVerbs are the prefixes that earn a command a short alias: `hub user list`
// reads better than `hub user list-user`, but only when stripping the noun
// leaves a sensible verb behind.
var crudVerbs = []string{"List", "Create", "Update", "Delete", "Get"}

// GroupName is the command a service is grouped under: "user.v1.UserService"
// becomes "user", "system.group.v1.GroupService" becomes "group".
func GroupName(op apicatalog.Operation) string {
	return kebab(strings.TrimSuffix(op.ServiceName, "Service"))
}

// CommandName is the subcommand for an rpc, the kebab-case of its name. It is
// deliberately mechanical: an agent that has read the rpc name in the proto or
// in `hub api list` can predict the command without consulting a mapping table.
func CommandName(op apicatalog.Operation) string {
	return kebab(op.Method)
}

// CommandAliases returns the shorthand for the plain CRUD rpcs - "list" for
// ListUser under `hub user` - and nothing for the rest.
func CommandAliases(op apicatalog.Operation) []string {
	entity := strings.TrimSuffix(op.ServiceName, "Service")
	for _, verb := range crudVerbs {
		if op.Method == verb+entity {
			return []string{strings.ToLower(verb)}
		}
	}
	return nil
}

// FlagName is the flag for a request field: "group_ids" becomes "group-ids".
func FlagName(field apicatalog.Field) string {
	return strings.ReplaceAll(field.Name, "_", "-")
}

// kebab lowercases a CamelCase identifier into dash-separated words, keeping
// runs of capitals together: "AddUsersToGroup" becomes "add-users-to-group" and
// "ListAPIKey" becomes "list-api-key".
func kebab(name string) string {
	var (
		out   strings.Builder
		runes = []rune(name)
	)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			previousIsLower := unicode.IsLower(runes[i-1])
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if previousIsLower || nextIsLower {
				out.WriteRune('-')
			}
		}
		out.WriteRune(unicode.ToLower(r))
	}
	return out.String()
}
