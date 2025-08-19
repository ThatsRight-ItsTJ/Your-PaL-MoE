package model

import (
	"encoding/json"
	"time"
)

// Enhanced Token fields for BricksLLM-style functionality
type TokenEnhanced struct {
	Token

	// Cost tracking and limits
	CostLimitUSD   *float64   `json:"cost_limit_usd" gorm:"type:decimal(10,4)"`
	CostUsedUSD    float64    `json:"cost_used_usd" gorm:"type:decimal(10,4);default:0.00"`
	CostResetDate  *time.Time `json:"cost_reset_date"`

	// Usage quotas and rate limits  
	RateLimitRPM        *int `json:"rate_limit_rpm"`
	RateLimitRPH        *int `json:"rate_limit_rph"`
	RateLimitRPD        *int `json:"rate_limit_rpd"`
	QuotaRequestsDaily  *int `json:"quota_requests_daily"`
	QuotaRequestsMonthly *int `json:"quota_requests_monthly"`
	RequestsUsedDaily    int  `json:"requests_used_daily" gorm:"default:0"`
	RequestsUsedMonthly  int  `json:"requests_used_monthly" gorm:"default:0"`
	LastRequestTime      *time.Time `json:"last_request_time"`

	// Tag-based organization
	Tags        json.RawMessage `json:"tags" gorm:"type:jsonb;default:'[]'"`
	Description string          `json:"description" gorm:"type:text"`
	Environment string          `json:"environment" gorm:"default:'production'"`

	// Key lifecycle management
	ExpiresAt             *time.Time `json:"expires_at"`
	AutoRotate            bool       `json:"auto_rotate" gorm:"default:false"`
	RotationIntervalDays  *int       `json:"rotation_interval_days"`
	LastRotatedAt         *time.Time `json:"last_rotated_at"`
	TokenStatus           string     `json:"token_status" gorm:"default:'active'"`

	// Model access control
	AllowedModels    json.RawMessage `json:"allowed_models" gorm:"type:jsonb;default:'[]'"`
	BlockedModels    json.RawMessage `json:"blocked_models" gorm:"type:jsonb;default:'[]'"`
	AllowedEndpoints json.RawMessage `json:"allowed_endpoints" gorm:"type:jsonb;default:'[]'"`

	// Enhanced metadata
	CreatedBy    *int       `json:"created_by" gorm:"index"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	UsageCount   int        `json:"usage_count" gorm:"default:0"`
}

// TableName returns the table name for TokenEnhanced
func (TokenEnhanced) TableName() string {
	return "tokens"
}