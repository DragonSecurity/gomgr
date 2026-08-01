package version

import "testing"

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()
	if info.Version == "" {
		t.Error("expected non-empty Version field")
	}
}

// A released binary is stamped with its tag at link time; that stamp has to
// win over the VCS revision, which is only ever a commit hash.
func TestGetBuildInfoPrefersStampedVersion(t *testing.T) {
	prev := stamped
	t.Cleanup(func() { stamped = prev })

	stamped = "v1.2.3"
	if got := GetBuildInfo().Version; got != "v1.2.3" {
		t.Errorf("Version = %q, want %q", got, "v1.2.3")
	}
}
