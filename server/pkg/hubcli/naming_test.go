package hubcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/server/pkg/apicatalog"
)

func TestCommandNaming(t *testing.T) {
	tests := []struct {
		ref     string
		group   string
		command string
		aliases []string
	}{
		{ref: "ListUser", group: "user", command: "list-user", aliases: []string{"list"}},
		{ref: "GetMe", group: "user", command: "get-me"},
		{ref: "SendMeVerifyEmail", group: "user", command: "send-me-verify-email"},
		{ref: "AddUsersToGroup", group: "group", command: "add-users-to-group"},
		{ref: "GetGroup", group: "group", command: "get-group", aliases: []string{"get"}},
		{ref: "ListMenuResource", group: "resource", command: "list-menu-resource"},
		{ref: "RemovePermissionsFromRole", group: "role", command: "remove-permissions-from-role"},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			op, err := apicatalog.Default().Find(tt.ref)
			require.NoError(t, err)

			assert.Equal(t, tt.group, GroupName(op))
			assert.Equal(t, tt.command, CommandName(op))
			assert.Equal(t, tt.aliases, CommandAliases(op))
		})
	}
}

func TestFlagName(t *testing.T) {
	assert.Equal(t, "group-ids", FlagName(apicatalog.Field{Name: "group_ids"}))
	assert.Equal(t, "id", FlagName(apicatalog.Field{Name: "id"}))
}

func TestKebab(t *testing.T) {
	tests := map[string]string{
		"":                  "",
		"User":              "user",
		"ListUser":          "list-user",
		"AddUsersToGroup":   "add-users-to-group",
		"ListAPIKey":        "list-api-key",
		"OIDCClient":        "oidc-client",
		"listUser":          "list-user",
		"SendMeVerifyEmail": "send-me-verify-email",
	}

	for input, want := range tests {
		assert.Equal(t, want, kebab(input), input)
	}
}
