package crit

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/checkpoint-restore/go-criu/v8/crit/images/stats"
)

const (
	StatsDump    = "stats-dump"
	StatsRestore = "stats-restore"
)

// Helper function to load stats file into Go struct
func getStats(path string) (*stats.StatsEntry, error) {
	statsFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer statsFile.Close()

	c := New(statsFile, nil, "", false, false)
	statsImg, err := c.Decode(&stats.StatsEntry{})
	if err != nil {
		return nil, err
	}

	stats, ok := statsImg.Entries[0].Message.(*stats.StatsEntry)
	if !ok {
		return nil, errors.New("failed to type assert stats image")
	}

	return stats, nil
}

// GetDumpStats returns the dump statistics of a checkpoint.
// dir is the path to the directory with the checkpoint images.
func GetDumpStats(dir string) (*stats.DumpStatsEntry, error) {
	stats, err := getStats(filepath.Join(dir, StatsDump))
	if err != nil {
		return nil, err
	}

	return stats.GetDump(), nil
}

// GetRestoreStats returns the restore statistics of a checkpoint.
// dir is the path to the directory with the checkpoint images.
func GetRestoreStats(dir string) (*stats.RestoreStatsEntry, error) {
	stats, err := getStats(filepath.Join(dir, StatsRestore))
	if err != nil {
		return nil, err
	}

	return stats.GetRestore(), nil
}
