package version

import "testing"

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()
	if info.Version == "" {
		t.Error("expected non-empty Version field")
	}
}
