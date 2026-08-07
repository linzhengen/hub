package di_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/linzhengen/hub/v1/server/config"
	"github.com/linzhengen/hub/v1/server/di"
	cmdregister "github.com/linzhengen/hub/v1/server/internal/interface/cmd/register"
	grpcregister "github.com/linzhengen/hub/v1/server/internal/interface/grpc/register"
)

// The container resolves lazily, so a constructor whose dependencies are not
// registered only fails when the process starts - after a deploy, not in CI.
// Building both entry points' object graphs here turns that into a test
// failure.
func TestContainerResolvesBothEntryPoints(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	container := di.NewDI(config.EnvConfig{Database: "postgres"}, db)

	t.Run("cli commands", func(t *testing.T) {
		require.NoError(t, container.Invoke(func(commands cmdregister.Commands) {
			require.NotEmpty(t, commands)
		}))
	})

	t.Run("grpc server", func(t *testing.T) {
		require.NoError(t, container.Invoke(func(opts *grpcregister.Opts) {
			require.NotNil(t, opts)
		}))
	})
}
