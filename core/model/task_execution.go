package model

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	ErrTaskExecutionNotFound    = "task_execution"
	ErrSubtaskExecutionNotFound = "subtask_execution"
)

const (
	TaskStatusPending   = "pending"
	TaskStatusExecuting = "executing"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

const (
	SubtaskTypeTextGeneration  = "text_generation"
	SubtaskTypeImageGeneration = "image_generation"
	SubtaskTypeAudioGeneration = "audio_generation"
	SubtaskTypeCodeGeneration  = "code_generation"
)

const (
	ProviderTierOfficial   = "official"
	ProviderTierCommunity  = "community"
	ProviderTierUnofficial = "unofficial"
)

type TaskExecution struct {
	ID          int             `json:"id" gorm:"primaryKey"`
	RequestID   string          `json:"request_id" gorm:"not null;index"`
	UserID      *int            `json:"user_id" gorm:"index"`
	User        *User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TokenID     *int            `json:"token_id" gorm:"index"`
	Token       *Token          `json:"token,omitempty" gorm:"foreignKey:TokenID"`

	// Original request info
	OriginalPrompt string `json:"original_prompt" gorm:"type:text;not null"`
	RequestType    string `json:"request_type" gorm:"not null"`

	// Task decomposition
	Subtasks       json.RawMessage `json:"subtasks" gorm:"type:jsonb;default:'[]'"`
	ExecutionPlan  json.RawMessage `json:"execution_plan" gorm:"type:jsonb;default:'{}'"`

	// Execution tracking
	Status           string     `json:"status" gorm:"default:'pending';index"`
	StartedAt        *time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	TotalDurationMs  *int       `json:"total_duration_ms"`

	// Cost and usage
	TotalCostUSD float64 `json:"total_cost_usd" gorm:"type:decimal(10,6);default:0.00"`
	TokensUsed   int     `json:"tokens_used" gorm:"default:0"`

	// Results
	FinalResponse json.RawMessage `json:"final_response" gorm:"type:jsonb"`
	ErrorMessage  string          `json:"error_message" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associations
	SubtaskExecutions []SubtaskExecution `json:"subtask_executions,omitempty" gorm:"foreignKey:TaskExecutionID"`
}

type SubtaskExecution struct {
	ID               int            `json:"id" gorm:"primaryKey"`
	TaskExecutionID  int            `json:"task_execution_id" gorm:"not null;index"`
	TaskExecution    *TaskExecution `json:"task_execution,omitempty" gorm:"foreignKey:TaskExecutionID"`
	SubtaskIndex     int            `json:"subtask_index" gorm:"not null"`

	// Subtask details
	SubtaskType  string `json:"subtask_type" gorm:"not null"`
	Prompt       string `json:"prompt" gorm:"type:text;not null"`
	ProviderName string `json:"provider_name" gorm:"not null;index"`
	ProviderTier string `json:"provider_tier" gorm:"not null"`
	ModelName    string `json:"model_name"`

	// Execution
	Status      string     `json:"status" gorm:"default:'pending'"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	DurationMs  *int       `json:"duration_ms"`
	RetryCount  int        `json:"retry_count" gorm:"default:0"`

	// Results and costs
	Response     json.RawMessage `json:"response" gorm:"type:jsonb"`
	CostUSD      float64         `json:"cost_usd" gorm:"type:decimal(10,6);default:0.00"`
	TokensUsed   int             `json:"tokens_used" gorm:"default:0"`
	ErrorMessage string          `json:"error_message" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task Execution functions
func CreateTaskExecution(te *TaskExecution) error {
	return DB.Create(te).Error
}

func GetTaskExecutionByID(id int) (*TaskExecution, error) {
	if id == 0 {
		return nil, errors.New("task execution id is empty")
	}

	var te TaskExecution
	err := DB.Preload("SubtaskExecutions").Where("id = ?", id).First(&te).Error

	return &te, HandleNotFound(err, ErrTaskExecutionNotFound)
}

func GetTaskExecutionByRequestID(requestID string) (*TaskExecution, error) {
	if requestID == "" {
		return nil, errors.New("request id is empty")
	}

	var te TaskExecution
	err := DB.Preload("SubtaskExecutions").Where("request_id = ?", requestID).First(&te).Error

	return &te, HandleNotFound(err, ErrTaskExecutionNotFound)
}

func UpdateTaskExecution(id int, updates map[string]interface{}) error {
	result := DB.Model(&TaskExecution{}).Where("id = ?", id).Updates(updates)
	return HandleUpdateResult(result, ErrTaskExecutionNotFound)
}

func GetTaskExecutions(userID *int, tokenID *int, status string, page, perPage int) ([]*TaskExecution, int64, error) {
	tx := DB.Model(&TaskExecution{})

	if userID != nil {
		tx = tx.Where("user_id = ?", *userID)
	}

	if tokenID != nil {
		tx = tx.Where("token_id = ?", *tokenID)
	}

	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var tasks []*TaskExecution
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("created_at desc").Limit(limit).Offset(offset).Find(&tasks).Error

	return tasks, total, err
}

func DeleteTaskExecution(id int) error {
	if id == 0 {
		return errors.New("task execution id is empty")
	}

	result := DB.Delete(&TaskExecution{ID: id})
	return HandleUpdateResult(result, ErrTaskExecutionNotFound)
}

// Subtask Execution functions
func CreateSubtaskExecution(se *SubtaskExecution) error {
	return DB.Create(se).Error
}

func GetSubtaskExecutionByID(id int) (*SubtaskExecution, error) {
	if id == 0 {
		return nil, errors.New("subtask execution id is empty")
	}

	var se SubtaskExecution
	err := DB.Where("id = ?", id).First(&se).Error

	return &se, HandleNotFound(err, ErrSubtaskExecutionNotFound)
}

func UpdateSubtaskExecution(id int, updates map[string]interface{}) error {
	result := DB.Model(&SubtaskExecution{}).Where("id = ?", id).Updates(updates)
	return HandleUpdateResult(result, ErrSubtaskExecutionNotFound)
}

func GetSubtaskExecutionsByTaskID(taskID int) ([]*SubtaskExecution, error) {
	var subtasks []*SubtaskExecution
	err := DB.Where("task_execution_id = ?", taskID).Order("subtask_index asc").Find(&subtasks).Error
	return subtasks, err
}

func GetSubtaskExecutionsByProvider(providerName string, page, perPage int) ([]*SubtaskExecution, int64, error) {
	tx := DB.Model(&SubtaskExecution{}).Where("provider_name = ?", providerName)

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var subtasks []*SubtaskExecution
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("created_at desc").Limit(limit).Offset(offset).Find(&subtasks).Error

	return subtasks, total, err
}

func DeleteSubtaskExecution(id int) error {
	if id == 0 {
		return errors.New("subtask execution id is empty")
	}

	result := DB.Delete(&SubtaskExecution{ID: id})
	return HandleUpdateResult(result, ErrSubtaskExecutionNotFound)
}

func SearchTaskExecutions(keyword string, userID *int, status string, page, perPage int) ([]*TaskExecution, int64, error) {
	tx := DB.Model(&TaskExecution{})

	if userID != nil {
		tx = tx.Where("user_id = ?", *userID)
	}

	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	if keyword != "" {
		tx = tx.Where("request_id LIKE ? OR original_prompt LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var tasks []*TaskExecution
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("created_at desc").Limit(limit).Offset(offset).Find(&tasks).Error

	return tasks, total, err
}