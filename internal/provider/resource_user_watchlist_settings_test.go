package provider

import "testing"

func TestUserWatchlistPathUsesMainSettingsEndpoint(t *testing.T) {
	if got, want := userWatchlistPath(1), "/api/v1/user/1/settings/main"; got != want {
		t.Fatalf("userWatchlistPath(1) = %q, want %q", got, want)
	}
}
