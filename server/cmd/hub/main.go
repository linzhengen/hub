// Command hub is the command line client for the hub API.
//
// It is a plain HTTP client: unlike cmd/cli, which runs operational tasks
// against the database, hub only ever talks to the REST gateway, so it needs a
// token and an endpoint and nothing else.
package main

import (
	"os"

	"github.com/linzhengen/hub/v1/server/pkg/hubcli"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(hubcli.Execute(hubcli.Options{Version: version}))
}
