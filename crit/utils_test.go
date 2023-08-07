package crit

import (
	"testing"
)

func TestFindPs(t *testing.T) {
	root := &PsTree{
		PID: 1,
		Children: []*PsTree{
			{
				PID: 2,
				Children: []*PsTree{
					{
						PID: 3,
					},
				},
			},
			{
				PID: 4,
				Children: []*PsTree{
					{
						PID: 5,
					},
				},
			},
		},
	}

	// Test Case 1: Find an existing process with a valid PID
	ps := root.FindPs(3)
	if ps == nil {
		t.Errorf("FindPs(3) returned nil, expected a valid process")
	}
	if ps != nil && ps.PID != 3 {
		t.Errorf("FindPs(3) returned a process with PID %d, expected 3", ps.PID)
	}

	// Test Case 2: Find a non-existing process with an invalid PID
	nonExistentPID := uint32(999)
	notFoundProcess := root.FindPs(nonExistentPID)
	if notFoundProcess != nil {
		t.Errorf("FindPs(%d) returned a process, expected nil", nonExistentPID)
	}
}

func TestCountBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{1000, "1000 B"},
		{5120, "5.0 KB"},
		{100000, "97.7 KB"},
	}

	for _, test := range tests {
		got := countBytes(test.input)
		if got != test.want {
			t.Errorf("want: %s, got: %s", test.want, got)
		}
	}
}

func TestProcessIP(t *testing.T) {
	tests := []struct {
		input []uint32
		want  string
	}{
		{[]uint32{}, ""},
		{[]uint32{0}, "0.0.0.0"},
		{[]uint32{16777343}, "127.0.0.1"},
		{[]uint32{0, 0, 0, 0}, "::"},
		{[]uint32{0, 0, 4294901760, 16777343}, "7f00:1::"},
	}

	for _, test := range tests {
		got := processIP(test.input)
		if got != test.want {
			t.Errorf("want: %s, got: %s", test.want, got)
		}
	}
}
