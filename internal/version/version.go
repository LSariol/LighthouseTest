package version

import (
	_ "embed"
	"os"
	"strings"
)

//go:embed VERSION
var embeddedVersionRaw string

var cached string

// Get returns the application version.
// Priority: env APP_VERSION -> embedded VERSION -> "dev".
func Get() string {
	if cached != "" {
		return cached
	}
	if v := strings.TrimSpace(os.Getenv("APP_VERSION")); v != "" {
		cached = v
		return cached
	}
	if v := strings.TrimSpace(embeddedVersionRaw); v != "" {
		cached = v
		return cached
	}
	cached = "dev"
	return cached
}
