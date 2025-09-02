package taskmaster

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/types"
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/pkg/selection"
	"github.com/sirupsen/logrus"
)

type TaskMaster struct {
	selector    *selection.EnhancedAdaptiveSelector
	activeJobs  map[string]*JobExecution
	jobsMutex   sync.RWMutex
	logger      *logrus.Logger
	maxWorkers  int
	workerPool  chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

type JobExecution struct {
	ID          string
	Status      JobStatus
	StartTime   time.Time
	EndTime     *time.Time
	Result      interface{}
	Error       error
	Progress    float64
	Metadata    map[string]interface{}
	mutex       sync.RWMutex
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type JobRequest struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Input    types.RequestInput     `json:"input"`
	Priority int                    `json:"priority"`
	Metadata map[string]interface{} `json:"metadata"`
}

type JobResult struct {
	ID       string      `json:"id"`
	Status   JobStatus   `json:"status"`
	Result   interface{} `json:"result,omitempty"`
	Error    string      `json:"error,omitempty"`
	Progress float64     `json:"progress"`
	Duration string      `json:"duration,omitempty"`
}

func NewTaskMaster(selector *selection.EnhancedAdaptiveSelector, logger *logrus.Logger, maxWorkers int) *TaskMaster {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &TaskMaster{
		selector:   selector,
		activeJobs: make(map[string]*JobExecution),
		logger:     logger,
		maxWorkers: maxWorkers,
		workerPool: make(chan struct{}, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (tm *TaskMaster) Start() error {
	tm.logger.Info("TaskMaster starting...")
	
	// Initialize worker pool
	for i := 0; i < tm.maxWorkers; i++ {
		tm.workerPool <- struct{}{}
	}
	
	tm.logger.Infof("TaskMaster started with %d workers", tm.maxWorkers)
	return nil
}

func (tm *TaskMaster) Stop() error {
	tm.logger.Info("TaskMaster stopping...")
	tm.cancel()
	
	// Wait for active jobs to complete or timeout
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()
	
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout.C:
			tm.logger.Warn("TaskMaster stop timeout reached, forcing shutdown")
			return nil
		case <-ticker.C:
			if tm.getActiveJobCount() == 0 {
				tm.logger.Info("TaskMaster stopped")
				return nil
			}
		}
	}
}

func (tm *TaskMaster) SubmitJob(req JobRequest) (*JobResult, error) {
	tm.logger.Infof("Submitting job %s of type %s", req.ID, req.Type)
	
	// Create job execution
	job := &JobExecution{
		ID:        req.ID,
		Status:    JobStatusPending,
		StartTime: time.Now(),
		Metadata:  req.Metadata,
	}
	
	// Store job
	tm.jobsMutex.Lock()
	tm.activeJobs[req.ID] = job
	tm.jobsMutex.Unlock()
	
	// Execute job asynchronously
	go tm.executeJob(job, req)
	
	return &JobResult{
		ID:       req.ID,
		Status:   JobStatusPending,
		Progress: 0.0,
	}, nil
}

func (tm *TaskMaster) executeJob(job *JobExecution, req JobRequest) {
	// Wait for worker
	<-tm.workerPool
	defer func() {
		tm.workerPool <- struct{}{}
	}()
	
	// Update status
	job.mutex.Lock()
	job.Status = JobStatusRunning
	job.mutex.Unlock()
	
	tm.logger.Infof("Executing job %s", job.ID)
	
	// Execute the actual job
	var err error
	
	// Simulate work for now - replace with actual job execution logic
	time.Sleep(2 * time.Second)
	
	// Update final status
	job.mutex.Lock()
	defer job.mutex.Unlock()
	
	endTime := time.Now()
	job.EndTime = &endTime
	
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err
		tm.logger.Errorf("Job %s failed: %v", job.ID, err)
	} else {
		job.Status = JobStatusCompleted
		job.Progress = 100.0
		tm.logger.Infof("Job %s completed", job.ID)
	}
}

func (tm *TaskMaster) GetJobStatus(jobID string) (*JobResult, error) {
	tm.jobsMutex.RLock()
	job, exists := tm.activeJobs[jobID]
	tm.jobsMutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}
	
	job.mutex.RLock()
	defer job.mutex.RUnlock()
	
	result := &JobResult{
		ID:       job.ID,
		Status:   job.Status,
		Result:   job.Result,
		Progress: job.Progress,
	}
	
	if job.Error != nil {
		result.Error = job.Error.Error()
	}
	
	if job.EndTime != nil {
		result.Duration = job.EndTime.Sub(job.StartTime).String()
	}
	
	return result, nil
}

func (tm *TaskMaster) ListJobs() ([]*JobResult, error) {
	tm.jobsMutex.RLock()
	defer tm.jobsMutex.RUnlock()
	
	results := make([]*JobResult, 0, len(tm.activeJobs))
	
	for _, job := range tm.activeJobs {
		job.mutex.RLock()
		result := &JobResult{
			ID:       job.ID,
			Status:   job.Status,
			Result:   job.Result,
			Progress: job.Progress,
		}
		
		if job.Error != nil {
			result.Error = job.Error.Error()
		}
		
		if job.EndTime != nil {
			result.Duration = job.EndTime.Sub(job.StartTime).String()
		}
		
		results = append(results, result)
		job.mutex.RUnlock()
	}
	
	return results, nil
}

func (tm *TaskMaster) CancelJob(jobID string) error {
	tm.jobsMutex.Lock()
	job, exists := tm.activeJobs[jobID]
	if exists {
		job.mutex.Lock()
		if job.Status == JobStatusPending || job.Status == JobStatusRunning {
			job.Status = JobStatusCancelled
			endTime := time.Now()
			job.EndTime = &endTime
		}
		job.mutex.Unlock()
	}
	tm.jobsMutex.Unlock()
	
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	tm.logger.Infof("Job %s cancelled", jobID)
	return nil
}

func (tm *TaskMaster) getActiveJobCount() int {
	tm.jobsMutex.RLock()
	defer tm.jobsMutex.RUnlock()
	
	count := 0
	for _, job := range tm.activeJobs {
		job.mutex.RLock()
		if job.Status == JobStatusPending || job.Status == JobStatusRunning {
			count++
		}
		job.mutex.RUnlock()
	}
	
	return count
}

func (tm *TaskMaster) CleanupCompletedJobs() {
	tm.jobsMutex.Lock()
	defer tm.jobsMutex.Unlock()
	
	for id, job := range tm.activeJobs {
		job.mutex.RLock()
		isCompleted := job.Status == JobStatusCompleted || 
		              job.Status == JobStatusFailed || 
		              job.Status == JobStatusCancelled
		
		// Clean up jobs older than 1 hour
		if isCompleted && job.EndTime != nil && 
		   time.Since(*job.EndTime) > time.Hour {
			delete(tm.activeJobs, id)
			tm.logger.Debugf("Cleaned up job %s", id)
		}
		job.mutex.RUnlock()
	}
}

// Helper function to determine task complexity based on input
func (tm *TaskMaster) analyzeTaskComplexity(input string) string {
	if strings.Contains(input, "complex") || strings.Contains(input, "advanced") {
		return "high"
	}
	if strings.Contains(input, "simple") || strings.Contains(input, "basic") {
		return "low"
	}
	return "medium"
}

// GetMetrics returns current TaskMaster metrics
func (tm *TaskMaster) GetMetrics() map[string]interface{} {
	tm.jobsMutex.RLock()
	defer tm.jobsMutex.RUnlock()
	
	pending := 0
	running := 0
	completed := 0
	failed := 0
	
	for _, job := range tm.activeJobs {
		job.mutex.RLock()
		switch job.Status {
		case JobStatusPending:
			pending++
		case JobStatusRunning:
			running++
		case JobStatusCompleted:
			completed++
		case JobStatusFailed:
			failed++
		}
		job.mutex.RUnlock()
	}
	
	return map[string]interface{}{
		"total_jobs":     len(tm.activeJobs),
		"pending_jobs":   pending,
		"running_jobs":   running,
		"completed_jobs": completed,
		"failed_jobs":    failed,
		"max_workers":    tm.maxWorkers,
		"available_workers": len(tm.workerPool),
	}
}