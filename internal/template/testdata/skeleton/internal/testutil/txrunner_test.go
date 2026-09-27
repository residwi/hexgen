package testutil_test

import (
	"__MODULE__/internal/platform/database"
	"__MODULE__/internal/testutil"
)

// Here rather than in txrunner.go because package testutil cannot import
// platform/database without cycling -- its tests import testutil. An external
// test file compiles as its own package, so it can.
var _ database.TxRunner = testutil.FakeTxRunner{}
