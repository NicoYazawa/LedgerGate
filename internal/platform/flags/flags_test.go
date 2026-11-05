package flags

import (
	"context"
	"testing"
)

func TestFlagDefaultsOff(t *testing.T) {
	t.Setenv("LG_FEATURE_NEW_GATEWAY", "")

	if FromEnv().Enabled(context.Background(), "new-gateway") {
		t.Fatal("Enabled() = true, want false by default")
	}
}

func TestFlagEnabledByEnvironment(t *testing.T) {
	t.Setenv("LG_FEATURE_NEW_GATEWAY", "true")

	if !FromEnv().Enabled(context.Background(), "new_gateway") {
		t.Fatal("Enabled() = false, want true from environment")
	}
}

func TestFlagOverrideCanDisableEnabledEnvironmentFlag(t *testing.T) {
	t.Setenv("LG_FEATURE_NEW_GATEWAY", "true")
	flags := WithOverrides(map[string]bool{"new-gateway": false})

	if flags.Enabled(context.Background(), "new-gateway") {
		t.Fatal("Enabled() = true, want false from override")
	}
}
