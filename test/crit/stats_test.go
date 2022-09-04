package main

import (
	"testing"

	"github.com/checkpoint-restore/go-criu/v6/crit"
)

func TestGetDumpStats(t *testing.T) {
	dumpStats, err := crit.GetDumpStats("test-imgs")
	if err != nil {
		t.Error("Failed to get stats")
	}
	if dumpStats.GetPagesWritten() == 0 {
		t.Error("PagesWritten is 0")
	}
}

func TestGetRestoreStats(t *testing.T) {
	restoreStats, err := crit.GetRestoreStats("test-imgs")
	if err != nil {
		t.Error("Failed to get stats")
	}
	if restoreStats.GetForkingTime() == 0 {
		t.Error("ForkingTime is 0")
	}
}
