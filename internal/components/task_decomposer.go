package components

// TaskDecomposer provides a placeholder for decomposing tasks into subtasks.
// Phase 1 minimal implementation to stabilize the surface.
type TaskDecomposer struct{}

// NewTaskDecomposer returns a new TaskDecomposer instance.
func NewTaskDecomposer() *TaskDecomposer {
	return &TaskDecomposer{}
}

// DecomposeTask splits a task description into a slice of subtasks.
// For Phase 1, return a single-element slice containing the original input if non-empty.
func (td *TaskDecomposer) DecomposeTask(input string) []string {
	if input == "" {
		return []string{}
	}
	return []string{input}
}
