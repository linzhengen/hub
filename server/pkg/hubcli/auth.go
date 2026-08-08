package hubcli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
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
		web      bool
		save     bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with hub",
		Long: `Authenticate with hub and store a token in the configuration profile.

Without flags, the browser-based device flow is used: a URL and a short code
are printed, the browser is opened automatically, and the CLI waits for you to
sign in. No client secret is required.

With --username / --password the OIDC password grant is used instead.
Without --username the client credentials grant is used, which is how an
unattended agent authenticates as a service account (requires a client secret).

The issuer and client come from the profile or from HUB_OIDC_ISSUER and
HUB_OIDC_CLIENT_ID. HUB_OIDC_CLIENT_SECRET is only needed for the client
credentials grant.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Default to web (device flow) when no credentials are given.
			useWeb := web || (username == "" && c.settings.OIDC.ClientSecret == "")

			var (
				token Token
				err   error
			)
			if useWeb {
				token, err = deviceFlow(cmd.Context(), cmd, c.settings.OIDC)
			} else {
				if username != "" {
					if password, err = readPassword(cmd, password); err != nil {
						return err
					}
				}
				token, err = requestToken(cmd.Context(), c.settings.OIDC, username, password)
			}
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
	cmd.Flags().BoolVar(&web, "web", false, "open a browser to authenticate (device flow, no client secret needed)")
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

// deviceAuthResponse is the response from the device authorization endpoint
// (RFC 8628 §3.2).
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// deviceFlow implements the OAuth 2.0 Device Authorization Grant (RFC 8628).
// It prints a URL and a short code for the user to enter in a browser, opens
// the browser automatically, then polls the token endpoint until the user
// approves or the code expires.
func deviceFlow(ctx context.Context, cmd *cobra.Command, oidc OIDC) (Token, error) {
	if oidc.Issuer == "" {
		return Token{}, fmt.Errorf("no OIDC issuer configured: set it in the profile or via HUB_OIDC_ISSUER")
	}
	if oidc.ClientID == "" {
		return Token{}, fmt.Errorf("no OIDC client configured: set it in the profile or via HUB_OIDC_CLIENT_ID")
	}

	// Step 1: Request a device code.
	form := url.Values{}
	form.Set("client_id", oidc.ClientID)
	if oidc.ClientSecret != "" {
		form.Set("client_secret", oidc.ClientSecret)
	}

	deviceEndpoint := strings.TrimRight(oidc.Issuer, "/") + "/protocol/openid-connect/auth/device"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
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

	var dar deviceAuthResponse
	if err := json.Unmarshal(raw, &dar); err != nil {
		return Token{}, fmt.Errorf("device authorization endpoint returned an unexpected body: %w", err)
	}
	if dar.DeviceCode == "" || dar.UserCode == "" {
		return Token{}, fmt.Errorf("device authorization endpoint returned an incomplete response")
	}

	// Step 2: Show the code and URL to the user.
	browserURL := dar.VerificationURIComplete
	if browserURL == "" {
		browserURL = dar.VerificationURI
	}
	out := cmd.ErrOrStderr()
	_, _ = fmt.Fprintf(out, "! First copy your one-time code: %s\n", dar.UserCode)
	_, _ = fmt.Fprintf(out, "- Press Enter to open %s in your browser... ", dar.VerificationURI)

	// Wait for Enter, then open the browser. If stdin is not a terminal (CI),
	// skip the prompt and just print the URL.
	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Consume the Enter key press.
		buf := make([]byte, 1)
		_, _ = os.Stdin.Read(buf)
	} else {
		_, _ = fmt.Fprintln(out)
	}
	openBrowser(browserURL)

	// Step 3: Poll the token endpoint.
	pollInterval := time.Duration(dar.Interval) * time.Second
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(dar.ExpiresIn) * time.Second)

	tokenEndpoint := strings.TrimRight(oidc.Issuer, "/") + "/protocol/openid-connect/token"
	for {
		if time.Now().After(deadline) {
			return Token{}, fmt.Errorf("device code expired before the user authenticated")
		}

		select {
		case <-ctx.Done():
			return Token{}, ctx.Err()
		case <-time.After(pollInterval):
		}

		pollForm := url.Values{}
		pollForm.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		pollForm.Set("device_code", dar.DeviceCode)
		pollForm.Set("client_id", oidc.ClientID)
		if oidc.ClientSecret != "" {
			pollForm.Set("client_secret", oidc.ClientSecret)
		}

		pollReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(pollForm.Encode()))
		if err != nil {
			return Token{}, err
		}
		pollReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		pollResp, err := httpClient.Do(pollReq)
		if err != nil {
			return Token{}, err
		}
		pollRaw, err := io.ReadAll(pollResp.Body)
		_ = pollResp.Body.Close()
		if err != nil {
			return Token{}, err
		}

		if pollResp.StatusCode == http.StatusOK {
			var token Token
			if err := json.Unmarshal(pollRaw, &token); err != nil {
				return Token{}, fmt.Errorf("token endpoint returned an unexpected body: %w", err)
			}
			if token.AccessToken == "" {
				return Token{}, fmt.Errorf("token endpoint returned no access_token")
			}
			if token.ExpiresIn > 0 {
				token.expiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
			}
			_, _ = fmt.Fprintln(out, "✓ Authentication successful")
			return token, nil
		}

		// Check for well-known error codes (RFC 8628 §3.5).
		var pollErr struct {
			Error string `json:"error"`
		}
		if jsonErr := json.Unmarshal(pollRaw, &pollErr); jsonErr == nil {
			switch pollErr.Error {
			case "authorization_pending":
				// The user hasn't approved yet — keep polling.
				continue
			case "slow_down":
				// The server wants us to back off.
				pollInterval += 5 * time.Second
				continue
			case "access_denied":
				return Token{}, fmt.Errorf("access denied: the user rejected the device authorization request")
			case "expired_token":
				return Token{}, fmt.Errorf("device code expired before the user authenticated")
			}
		}
		return Token{}, &APIError{StatusCode: pollResp.StatusCode, Status: pollResp.Status, Body: trimJSON(pollRaw)}
	}
}

// openBrowser attempts to open url in the default system browser.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
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
