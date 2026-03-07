package client

type Client struct {
	ClientId      string
	ClientSecret  string
	RedirectUri   string
	GrantTypes    map[string]interface{}
	ResponseTypes map[string]interface{}
	Scopes        map[string]interface{}
}
