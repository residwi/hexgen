package scaffold

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDrop(t *testing.T) {
	cases := map[string]bool{
		// kept
		"cmd/api/main.go":                                false,
		"internal/config/config.go":                      false,
		"internal/core/slugify.go":                       false,
		"internal/core/apperror.go":                      false,
		"internal/core/response/error_mapper.go":         false,
		"internal/features/auth/service.go":              false,
		"internal/features/user/repository.go":           false,
		"internal/platform/database/postgres.go":         false,
		"internal/platform/jobs/runner.go":               false,
		"mocks/auth/mock_UserProvider.go":                false,
		"mocks/user/mock_Repository.go":                  false,
		"mocks/middleware/mock_TokenValidator.go":        false,
		"db/migrations/20260424120000_util_triggers.sql": false,
		"db/migrations/20260424120001_create_users.sql":  false,
		"bin/.keep":     false,
		".golangci.yml": false,
		// dropped
		"AGENTS.md":                                          true,
		".air.worker.toml":                                   true,
		"cmd/worker/main.go":                                 true,
		"internal/platform/payment/gateway.go":               true,
		"internal/wiring/order.go":                           true,
		"internal/core/money.go":                             true,
		"internal/core/money_test.go":                        true,
		"internal/core/address.go":                           true,
		"internal/platform/email/sender.go":                  true,
		"internal/platform/storage/s3.go":                    true,
		"internal/features/cart/service.go":                  true,
		"internal/features/order/repository.go":              true,
		"mocks/cart/mock_Repository.go":                      true,
		"mocks/payment/mock_Gateway.go":                      true,
		"db/migrations/20260424120002_create_categories.sql": true,
		"db/seeds/old.sql":                                   false, // seeds replaced by override, not dropped here
		"bin/run_test":                                       true,
	}
	for path, want := range cases {
		assert.Equalf(t, want, drop(path), "drop(%q)", path)
	}
}
