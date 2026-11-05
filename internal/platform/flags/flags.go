package flags

import (
	"context"
	"maps"
	"os"
	"strings"
)

type Flags interface {
	Enabled(ctx context.Context, name string) bool
}

type environmentFlags struct{}
type overrideFlags map[string]bool

func FromEnv() Flags { return environmentFlags{} }

func WithOverrides(values map[string]bool) Flags { return maps.Clone(overrideFlags(values)) }

func (environmentFlags) Enabled(_ context.Context, name string) bool {
	return enabledFromEnvironment(name)
}

func (values overrideFlags) Enabled(_ context.Context, name string) bool {
	if enabled, ok := values[name]; ok {
		return enabled
	}
	return enabledFromEnvironment(name)
}

func enabledFromEnvironment(name string) bool {
	value, ok := os.LookupEnv(environmentName(name))
	return ok && strings.EqualFold(strings.TrimSpace(value), "true")
}

func environmentName(name string) string {
	normalized := strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(strings.TrimSpace(name))
	return "LG_FEATURE_" + strings.ToUpper(normalized)
}
