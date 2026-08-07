package hubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func (c *CLI) newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Obtain and inspect API credentials",
	}
	cmd.AddCommand(c.newAuthLoginCommand(), c.newAuthTokenCommand(), c.newAuthWhoamiCommand())
	return cmd
}

func (c *CLI) newAuthLoginCommand() *cobra.Command {
	var (
		username string
		password string
		save     bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Exchange credentials for an access token",
		Long: `Exchange credentials for an access token.

With --username and --password the OIDC password grant is used; without them
the client credentials grant is, which is how an unattended agent authenticates
as a service account. The issuer and client come from the profile or from
HUB_OIDC_ISSUER, HUB_OIDC_CLIENT_ID and HUB_OIDC_CLIENT_SECRET.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := requestToken(cmd.Context(), c.settings.OIDC, username, password)
			if err != nil {
				return err
			}
			if save {
				if err := c.saveToken(token.AccessToken); err != nil {
					return err
				}
			}
			return Render(cmd.OutOrStdout(), c.format, token)
		},
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "username for the password grant")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password for the password grant")
	cmd.Flags().BoolVar(&save, "save", true, "store the token in the configuration profile")
	return cmd
}

func (c *CLI) saveToken(token string) error {
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

	profile := cfg.Profile(name)
	profile.Token = token
	if profile.Endpoint == "" {
		profile.Endpoint = c.settings.Endpoint
	}
	if profile.OIDC.Issuer == "" {
		profile.OIDC = c.settings.OIDC
	}
	return SaveConfig(path, cfg.WithProfile(name, profile))
}

func (c *CLI) newAuthTokenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print the access token in use",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if c.settings.Token == "" {
				return fmt.Errorf("no token configured: run `hub auth login`, set HUB_TOKEN or pass --token")
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), c.settings.Token)
			return err
		},
	}
}

func (c *CLI) newAuthWhoamiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the profile and groups of the authenticated caller",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			op, err := c.catalog.Find("/user.v1.UserService/GetMe")
			if err != nil {
				return err
			}
			return c.send(cmd, op, nil)
		},
	}
}

// Token is the subset of an OAuth2 token response the CLI needs.
type Token struct {
	AccessToken  string `json:"access_token" yaml:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty" yaml:"expires_in,omitempty"`
	TokenType    string `json:"token_type,omitempty" yaml:"token_type,omitempty"`
}

// requestToken runs an OIDC token request against the issuer. A username
// selects the password grant; its absence selects client credentials.
func requestToken(ctx context.Context, oidc OIDC, username, password string) (Token, error) {
	if oidc.Issuer == "" {
		return Token{}, fmt.Errorf("no OIDC issuer configured: set it in the profile or via HUB_OIDC_ISSUER")
	}
	if oidc.ClientID == "" {
		return Token{}, fmt.Errorf("no OIDC client configured: set it in the profile or via HUB_OIDC_CLIENT_ID")
	}

	form := url.Values{}
	form.Set("client_id", oidc.ClientID)
	if oidc.ClientSecret != "" {
		form.Set("client_secret", oidc.ClientSecret)
	}
	if username != "" {
		form.Set("grant_type", "password")
		form.Set("username", username)
		form.Set("password", password)
	} else {
		form.Set("grant_type", "client_credentials")
	}

	endpoint := strings.TrimRight(oidc.Issuer, "/") + "/protocol/openid-connect/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Token{}, &APIError{StatusCode: resp.StatusCode, Status: resp.Status, Body: trimJSON(raw)}
	}

	var token Token
	if err := json.Unmarshal(raw, &token); err != nil {
		return Token{}, fmt.Errorf("token endpoint returned an unexpected body: %w", err)
	}
	if token.AccessToken == "" {
		return Token{}, fmt.Errorf("token endpoint returned no access_token")
	}
	return token, nil
}
