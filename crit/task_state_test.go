package crit

import "testing"

func TestTaskStateString(t *testing.T) {
	tests := []struct {
		state TaskState
		want  string
	}{
		{TaskAlive, "Alive"},
		{TaskDead, "Dead"},
		{TaskStopped, "Stopped"},
		{TaskZombie, "Zombie"},
		{TaskState(0xFF), "Unknown"}, // Unknown state
	}

	for _, test := range tests {
		got := test.state.String()
		if got != test.want {
			t.Errorf("TaskState(%d).String() = %q, want %q", test.state, got, test.want)
		}
	}
}

func TestTaskStateIsAliveOrStopped(t *testing.T) {
	tests := []struct {
		state TaskState
		want  bool
	}{
		{TaskAlive, true},
		{TaskDead, false},
		{TaskStopped, true},
		{TaskZombie, false},
		{TaskState(0xFF), false}, // Unknown state
	}

	for _, test := range tests {
		got := test.state.IsAliveOrStopped()
		if got != test.want {
			t.Errorf("TaskState(%d).IsAliveOrStopped() = %v, want %v", test.state, got, test.want)
		}
	}
}

func TestTaskStateChecks(t *testing.T) {
	tests := []struct {
		state     TaskState
		isAlive   bool
		isDead    bool
		isStopped bool
		isZombie  bool
	}{
		{TaskAlive, true, false, false, false},
		{TaskDead, false, true, false, false},
		{TaskStopped, false, false, true, false},
		{TaskZombie, false, false, false, true},
		{TaskState(0xFF), false, false, false, false}, // Unknown
	}

	for _, test := range tests {
		if got := test.state.IsAlive(); got != test.isAlive {
			t.Errorf("TaskState(%d).IsAlive() = %v, want %v", test.state, got, test.isAlive)
		}
		if got := test.state.IsDead(); got != test.isDead {
			t.Errorf("TaskState(%d).IsDead() = %v, want %v", test.state, got, test.isDead)
		}
		if got := test.state.IsStopped(); got != test.isStopped {
			t.Errorf("TaskState(%d).IsStopped() = %v, want %v", test.state, got, test.isStopped)
		}
		if got := test.state.IsZombie(); got != test.isZombie {
			t.Errorf("TaskState(%d).IsZombie() = %v, want %v", test.state, got, test.isZombie)
		}
	}
}
