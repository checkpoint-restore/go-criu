package crit

type TaskState uint32

// Task state constants from CRIU (compel/include/uapi/task-state.h)
const (
	TaskAlive   TaskState = 0x01
	TaskDead    TaskState = 0x02
	TaskStopped TaskState = 0x03
	TaskZombie  TaskState = 0x06
)

func (s TaskState) String() string {
	if v, ok := map[TaskState]string{
		TaskAlive:   "Alive",
		TaskDead:    "Dead",
		TaskStopped: "Stopped",
		TaskZombie:  "Zombie",
	}[s]; ok {
		return v
	}
	return "Unknown"
}

// A checkpointed process has state (memory pages, file descriptors, sockets),
// only if it is "alive" or "stopped". Note that "dead" means the task has
// exited during checkpointing and CRIU observed the exit reported by waitpid().
func (s TaskState) IsAliveOrStopped() bool {
	switch s {
	case TaskAlive, TaskStopped:
		return true
	default:
		return false
	}
}

func (s TaskState) IsAlive() bool   { return s == TaskAlive }
func (s TaskState) IsDead() bool    { return s == TaskDead }
func (s TaskState) IsStopped() bool { return s == TaskStopped }
func (s TaskState) IsZombie() bool  { return s == TaskZombie }
