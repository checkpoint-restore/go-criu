package crit

import "testing"

func TestGetDumpStats(t *testing.T) {
	dumpStats, err := GetDumpStats("test-imgs")
	if err != nil {
		t.Error("Failed to get stats")
	}
	if dumpStats.GetPagesWritten() == 0 {
		t.Error("PagesWritten is 0")
	}
}

func TestGetRestoreStats(t *testing.T) {
	restoreStats, err := GetRestoreStats("test-imgs")
	if err != nil {
		t.Error("Failed to get stats")
	}
	if restoreStats.GetForkingTime() == 0 {
		t.Error("ForkingTime is 0")
	}
}
