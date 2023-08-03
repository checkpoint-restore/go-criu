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
