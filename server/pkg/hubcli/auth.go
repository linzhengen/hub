package hubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
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
			if username != "" {
				var err error
				if password, err = readPassword(cmd, password); err != nil {
					return err
				}
			}

			token, err := requestToken(cmd.Context(), c.settings.OIDC, username, password)
			if err != nil {
				return err
			}
			if save {
				if err := c.saveToken(token); err != nil {
					return err
				}
			}
			return Render(cmd.OutOrStdout(), c.format, token)
		},
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "username for the password grant")
	cmd.Flags().StringVarP(&password, "password", "p", "",
		"password for the password grant. Prefer the prompt or HUB_PASSWORD: a password given here is visible in the shell history and in the process list")
	cmd.Flags().BoolVar(&save, "save", true, "store the token in the configuration profile")
	return cmd
}

// readPassword resolves the password for the password grant without putting it
// on the command line where the shell would record it.
func readPassword(cmd *cobra.Command, flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if fromEnv := os.Getenv("HUB_PASSWORD"); fromEnv != "" {
		return fromEnv, nil
	}

	// Not a terminal - CI, or an agent driving the CLI - so there is nobody to
	// prompt. Say what to set rather than blocking on a read that never returns.
	stdin := int(os.Stdin.Fd())
	if !term.IsTerminal(stdin) {
		return "", fmt.Errorf("no password given: pass --password or set HUB_PASSWORD when stdin is not a terminal")
	}

	// The prompt goes to stderr so that stdout stays parseable.
	_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Password: ")
	entered, err := term.ReadPassword(stdin)
	// ReadPassword leaves the cursor on the prompt line either way.
	_, _ = fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return string(entered), nil
}

// saveToken writes a freshly issued token into the profile, keeping the refresh
// token and expiry so the next command can renew it without a second login.
func (c *CLI) saveToken(token Token) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		return err
	}

	name := c.profileName(cfg)
	profile := cfg.Profile(name)
	profile.Token = token.AccessToken
	profile.ExpiresAt = token.ExpiresAt()
	// A refresh grant may or may not rotate the refresh token; keep the old one
	// when the response omits it.
	if token.RefreshToken != "" {
		profile.RefreshToken = token.RefreshToken
	}
	if profile.Endpoint == "" {
		profile.Endpoint = c.settings.Endpoint
	}
	if profile.OIDC.Issuer == "" {
		profile.OIDC = c.settings.OIDC
	}
	return SaveConfig(path, cfg.WithProfile(name, profile))
}

// profileName is the profile a write should land in: the one named by
// --profile, else the configured current one, else "default".
func (c *CLI) profileName(cfg Config) string {
	if c.profile != "" {
		return c.profile
	}
	if cfg.Current != "" {
		return cfg.Current
	}
	return "default"
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

// refreshLeeway renews a token slightly before it expires, so a call cannot
// be refused because the token died between the check and the request.
const refreshLeeway = 30 * time.Second

// ensureFreshToken renews the access token when it is spent and a refresh token
// is available. Keycloak access tokens last minutes, so without this every
// command more than a few minutes after `hub auth login` would fail.
//
// A token supplied on the command line or through the environment is left
// alone: the caller chose it deliberately, and it is not ours to replace or to
// write into their config file.
func (c *CLI) ensureFreshToken(ctx context.Context) error {
	if c.tokenOverridden || c.settings.RefreshToken == "" {
		return nil
	}
	if c.settings.Token != "" && !c.tokenExpired() {
		return nil
	}

	token, err := refreshToken(ctx, c.settings.OIDC, c.settings.RefreshToken)
	if err != nil {
		return fmt.Errorf("the access token has expired and could not be renewed, run `hub auth login`: %w", err)
	}

	c.settings.Token = token.AccessToken
	c.settings.ExpiresAt = token.ExpiresAt()
	if token.RefreshToken != "" {
		c.settings.RefreshToken = token.RefreshToken
	}
	c.client = NewClient(c.settings)

	return c.saveToken(token)
}

// tokenExpired reports whether the access token is spent. An unknown expiry is
// treated as still valid: the provider did not say, so the only way to find out
// is to make the call.
func (c *CLI) tokenExpired() bool {
	if c.settings.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(refreshLeeway).After(c.settings.ExpiresAt)
}

// Token is the subset of an OAuth2 token response the CLI needs.
type Token struct {
	AccessToken  string `json:"access_token" yaml:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty" yaml:"expires_in,omitempty"`
	TokenType    string `json:"token_type,omitempty" yaml:"token_type,omitempty"`

	// expiresAt is when AccessToken stops being accepted, worked out from
	// ExpiresIn at the moment the response arrived. The wire format carries a
	// duration, which is useless once it has been written to a file.
	expiresAt time.Time
}

// ExpiresAt is when the access token stops being accepted, or the zero time if
// the provider did not say.
func (t Token) ExpiresAt() time.Time {
	return t.expiresAt
}

// requestToken runs an OIDC token request against the issuer. A username
// selects the password grant; its absence selects client credentials.
func requestToken(ctx context.Context, oidc OIDC, username, password string) (Token, error) {
	form := url.Values{}
	if username != "" {
		form.Set("grant_type", "password")
		form.Set("username", username)
		form.Set("password", password)
	} else {
		form.Set("grant_type", "client_credentials")
	}
	return exchange(ctx, oidc, form)
}

// refreshToken trades a refresh token for a new access token, which is how the
// CLI stays usable for longer than one access token's lifetime.
func refreshToken(ctx context.Context, oidc OIDC, refresh string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refresh)
	return exchange(ctx, oidc, form)
}

// exchange posts a grant to the issuer's token endpoint.
func exchange(ctx context.Context, oidc OIDC, form url.Values) (Token, error) {
	if oidc.Issuer == "" {
		return Token{}, fmt.Errorf("no OIDC issuer configured: set it in the profile or via HUB_OIDC_ISSUER")
	}
	if oidc.ClientID == "" {
		return Token{}, fmt.Errorf("no OIDC client configured: set it in the profile or via HUB_OIDC_CLIENT_ID")
	}

	form.Set("client_id", oidc.ClientID)
	if oidc.ClientSecret != "" {
		form.Set("client_secret", oidc.ClientSecret)
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
	if token.ExpiresIn > 0 {
		token.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	return token, nil
}
