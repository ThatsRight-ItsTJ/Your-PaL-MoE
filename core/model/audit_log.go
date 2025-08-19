package model

import (
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	ErrAuditLogNotFound = "audit_log"
)

const (
	AuditActionAPIRequest    = "api_request"
	AuditActionKeyCreated    = "key_created"
	AuditActionKeyRotated    = "key_rotated"
	AuditActionKeyDeleted    = "key_deleted"
	AuditActionKeyUpdated    = "key_updated"
	AuditActionUserCreated   = "user_created"
	AuditActionUserUpdated   = "user_updated"
	AuditActionUserDeleted   = "user_deleted"
	AuditActionGroupCreated  = "group_created"
	AuditActionGroupUpdated  = "group_updated"
	AuditActionGroupDeleted  = "group_deleted"
	AuditActionLogin         = "login"
	AuditActionLogout        = "logout"
	AuditActionPasswordReset = "password_reset"
)

const (
	ResourceTypeToken = "token"
	ResourceTypeUser  = "user"
	ResourceTypeGroup = "group"
	ResourceTypeAPI   = "api"
)

type AuditLog struct {
	ID         int    `json:"id" gorm:"primaryKey"`
	UserID     *int   `json:"user_id" gorm:"index"`
	User       *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	TokenID    *int   `json:"token_id" gorm:"index"`
	Token      *Token `json:"token,omitempty" gorm:"foreignKey:TokenID"`

	// Action details
	Action       string `json:"action" gorm:"not null;index"`
	ResourceType string `json:"resource_type" gorm:"not null"`
	ResourceID   string `json:"resource_id"`

	// Request context
	IPAddress  *net.IP `json:"ip_address" gorm:"type:inet"`
	UserAgent  string  `json:"user_agent" gorm:"type:text"`
	Endpoint   string  `json:"endpoint"`
	HTTPMethod string  `json:"http_method"`

	// Changes and metadata
	OldValues json.RawMessage `json:"old_values" gorm:"type:jsonb"`
	NewValues json.RawMessage `json:"new_values" gorm:"type:jsonb"`
	Metadata  json.RawMessage `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Results
	Success      bool   `json:"success" gorm:"default:true"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func CreateAuditLog(al *AuditLog) error {
	return DB.Create(al).Error
}

func GetAuditLogByID(id int) (*AuditLog, error) {
	if id == 0 {
		return nil, errors.New("audit log id is empty")
	}

	var al AuditLog
	err := DB.Preload("User").Preload("Token").Where("id = ?", id).First(&al).Error

	return &al, HandleNotFound(err, ErrAuditLogNotFound)
}

func GetAuditLogs(userID *int, tokenID *int, action, resourceType string, page, perPage int) ([]*AuditLog, int64, error) {
	tx := DB.Model(&AuditLog{})

	if userID != nil {
		tx = tx.Where("user_id = ?", *userID)
	}

	if tokenID != nil {
		tx = tx.Where("token_id = ?", *tokenID)
	}

	if action != "" {
		tx = tx.Where("action = ?", action)
	}

	if resourceType != "" {
		tx = tx.Where("resource_type = ?", resourceType)
	}

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var logs []*AuditLog
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, total, err
}

func SearchAuditLogs(keyword string, userID *int, action string, page, perPage int) ([]*AuditLog, int64, error) {
	tx := DB.Model(&AuditLog{})

	if userID != nil {
		tx = tx.Where("user_id = ?", *userID)
	}

	if action != "" {
		tx = tx.Where("action = ?", action)
	}

	if keyword != "" {
		tx = tx.Where("resource_id LIKE ? OR endpoint LIKE ? OR error_message LIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var logs []*AuditLog
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, total, err
}

func DeleteAuditLog(id int) error {
	if id == 0 {
		return errors.New("audit log id is empty")
	}

	result := DB.Delete(&AuditLog{ID: id})
	return HandleUpdateResult(result, ErrAuditLogNotFound)
}

func DeleteAuditLogsBefore(cutoffDate time.Time) (int64, error) {
	result := DB.Where("created_at < ?", cutoffDate).Delete(&AuditLog{})
	return result.RowsAffected, result.Error
}

func GetAuditLogsByDateRange(startDate, endDate time.Time, page, perPage int) ([]*AuditLog, int64, error) {
	tx := DB.Model(&AuditLog{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var logs []*AuditLog
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error

	return logs, total, err
}

func LogAuditEvent(userID *int, tokenID *int, action, resourceType, resourceID string, 
	ipAddr *net.IP, userAgent, endpoint, httpMethod string, 
	oldValues, newValues, metadata json.RawMessage, 
	success bool, errorMessage string) error {
	
	audit := &AuditLog{
		UserID:       userID,
		TokenID:      tokenID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ipAddr,
		UserAgent:    userAgent,
		Endpoint:     endpoint,
		HTTPMethod:   httpMethod,
		OldValues:    oldValues,
		NewValues:    newValues,
		Metadata:     metadata,
		Success:      success,
		ErrorMessage: errorMessage,
	}

	return CreateAuditLog(audit)
}

func GetAuditLogStats(startDate, endDate time.Time) (map[string]interface{}, error) {
	type ActionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}

	var actionCounts []ActionCount
	err := DB.Model(&AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("action").
		Find(&actionCounts).Error
	if err != nil {
		return nil, err
	}

	var totalLogs int64
	err = DB.Model(&AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalLogs).Error
	if err != nil {
		return nil, err
	}

	var successfulLogs int64
	err = DB.Model(&AuditLog{}).
		Where("created_at BETWEEN ? AND ? AND success = ?", startDate, endDate, true).
		Count(&successfulLogs).Error
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_logs":      totalLogs,
		"successful_logs": successfulLogs,
		"failed_logs":     totalLogs - successfulLogs,
		"action_counts":   actionCounts,
		"success_rate":    float64(successfulLogs) / float64(totalLogs) * 100,
	}

	return stats, nil
}