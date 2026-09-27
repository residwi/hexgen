// Package seeds proves db/seeds/data.sql still applies against a freshly
// migrated database. Nothing else in the suite touches that file, so a dropped
// or renamed column can break it without any test noticing.
package seeds

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"__MODULE__/internal/testutil"
)

func TestDataSQL_Applies(t *testing.T) {
	pool, cleanup := testutil.MustStartPostgres("test_db_seeds")
	defer cleanup()
	ctx := context.Background()

	seedSQL, err := os.ReadFile(seedFilePath())
	require.NoError(t, err)

	_, err = pool.Exec(ctx, string(seedSQL))
	require.NoError(t, err, "db/seeds/data.sql must apply cleanly against a freshly migrated database")

	// The seed must produce a real admin, not merely run without an SQL error.
	var role string
	var active bool
	err = pool.QueryRow(ctx,
		`SELECT role, active FROM users WHERE email = 'admin@example.com'`,
	).Scan(&role, &active)
	require.NoError(t, err, "seed must insert the dev admin user")
	assert.Equal(t, "admin", role)
	assert.True(t, active)

	// Re-applying must stay a no-op, as running `make seed` twice would.
	_, err = pool.Exec(ctx, string(seedSQL))
	require.NoError(t, err, "re-applying db/seeds/data.sql must also succeed")
}

// Resolved relative to this source file, as testutil locates the migrations
// directory.
func seedFilePath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "data.sql")
}
