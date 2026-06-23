package scaffold

import (
	pathpkg "path"
	"strings"
)

// drop reports whether a slash-separated template path should be excluded from
// the generated skeleton.
func drop(p string) bool {
	switch {
	case p == "AGENTS.md":
		return true
	case p == ".air.worker.toml":
		return true
	case p == "cmd/worker" || strings.HasPrefix(p, "cmd/worker/"):
		return true
	case p == "internal/platform/payment" || strings.HasPrefix(p, "internal/platform/payment/"):
		return true
	case p == "internal/wiring" || strings.HasPrefix(p, "internal/wiring/"):
		// Cross-feature adapters for the dropped e-commerce features; they import
		// cart/order/payment/etc. and would not compile in the skeleton.
		return true
	case p == "internal/core/money.go", p == "internal/core/money_test.go",
		p == "internal/core/address.go", p == "internal/core/address_test.go":
		// E-commerce value objects (Money, Address); unused by the auth+user skeleton.
		return true
	case strings.HasPrefix(p, "internal/platform/email/"),
		strings.HasPrefix(p, "internal/platform/storage/"):
		// Optional infra (email sender, object storage) not wired in the skeleton.
		return true
	case strings.HasPrefix(p, "internal/features/"):
		name := segment(p, 2)
		return name != "auth" && name != "user"
	case strings.HasPrefix(p, "mocks/"):
		name := segment(p, 1)
		return name != "auth" && name != "user" && name != "middleware"
	case strings.HasPrefix(p, "db/migrations/"):
		base := pathpkg.Base(p)
		return !strings.HasSuffix(base, "_util_triggers.sql") &&
			!strings.HasSuffix(base, "_create_users.sql")
	case strings.HasPrefix(p, "bin/"):
		return pathpkg.Base(p) != ".keep"
	default:
		return false
	}
}

// segment returns the n-th (0-based) slash segment of p, or "" if out of range.
func segment(p string, n int) string {
	for range n {
		i := strings.IndexByte(p, '/')
		if i < 0 {
			return ""
		}
		p = p[i+1:]
	}
	if before, _, ok := strings.Cut(p, "/"); ok {
		return before
	}
	return p
}
