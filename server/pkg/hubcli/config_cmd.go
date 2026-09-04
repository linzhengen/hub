package hubcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *CLI) newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and edit the CLI configuration",
	}
	cmd.AddCommand(c.newConfigPathCommand(), c.newConfigShowCommand(), c.newConfigSetCommand())
	return cmd
}

func (c *CLI) newConfigPathCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the configuration file location",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := ConfigPath()
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	}
}

func (c *CLI) newConfigShowCommand() *cobra.Command {
	var revealToken bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the settings the CLI resolved",
		Long: `Show the settings the CLI resolved from the profile, the environment and the
flags. The token is redacted unless --reveal-token is given, so the output is
safe to paste into an issue.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			settings := c.settings
			if !revealToken {
				settings.Token = redact(settings.Token)
				settings.OIDC.ClientSecret = redact(settings.OIDC.ClientSecret)
			}
			return Render(cmd.OutOrStdout(), c.format, map[string]any{
				"endpoint":         settings.Endpoint,
				"token":            settings.Token,
				"oidcIssuer":       settings.OIDC.Issuer,
				"oidcClientId":     settings.OIDC.ClientID,
				"oidcClientSecret": settings.OIDC.ClientSecret,
			})
		},
	}
	cmd.Flags().BoolVar(&revealToken, "reveal-token", false, "print the token and client secret in full")
	return cmd
}

func (c *CLI) newConfigSetCommand() *cobra.Command {
	var profile Profile
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Write connection settings into a profile",
		Long: `Write connection settings into a profile.

Only the flags given are changed; the rest of the profile is left alone.

  hub config set --endpoint http://localhost:9090 \
      --oidc-issuer http://localhost:8080/realms/hub --oidc-client-id hub-web`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := ConfigPath()
			if err != nil {
				return err
			}
			cfg, err := LoadConfig(path)
			if err != nil {
				return err
			}

			name := c.profile
			if name == "" {
				name = cfg.Current
			}
			if name == "" {
				name = "default"
			}

			existing := cfg.Profile(name)
			override(&existing.Endpoint, profile.Endpoint)
			override(&existing.Token, profile.Token)
			override(&existing.OIDC.Issuer, profile.OIDC.Issuer)
			override(&existing.OIDC.ClientID, profile.OIDC.ClientID)
			override(&existing.OIDC.ClientSecret, profile.OIDC.ClientSecret)

			if err := SaveConfig(path, cfg.WithProfile(name, existing)); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "updated profile %q in %s\n", name, path)
			return err
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&profile.Endpoint, "endpoint", "", "base URL of the hub API")
	flags.StringVar(&profile.Token, "token", "", "bearer token")
	flags.StringVar(&profile.OIDC.Issuer, "oidc-issuer", "", "OIDC realm URL")
	flags.StringVar(&profile.OIDC.ClientID, "oidc-client-id", "", "OIDC client id")
	flags.StringVar(&profile.OIDC.ClientSecret, "oidc-client-secret", "", "OIDC client secret")
	return cmd
}

func redact(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}
