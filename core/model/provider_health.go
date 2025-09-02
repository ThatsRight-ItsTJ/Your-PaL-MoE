package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	ErrProviderHealthNotFound = "provider_health"
)

const (
	ProviderStatusHealthy  = "healthy"
	ProviderStatusDegraded = "degraded"
	ProviderStatusDown     = "down"
	ProviderStatusUnknown  = "unknown"
)

const (
	ProviderTypeOfficial   = "official"
	ProviderTypeCommunity  = "community"
	ProviderTypeUnofficial = "unofficial"
)

type ProviderHealth struct {
	ID                   int       `json:"id" gorm:"primaryKey"`
	ProviderName         string    `json:"provider_name" gorm:"not null;index"`
	ProviderType         string    `json:"provider_type" gorm:"not null;index;default:'official'"`
	EndpointURL          string    `json:"endpoint_url" gorm:"not null"`
	Status               string    `json:"status" gorm:"default:'unknown';index"`
	ResponseTimeMs       *int      `json:"response_time_ms"`
	SuccessRate          *float64  `json:"success_rate" gorm:"type:decimal(5,2)"`
	LastCheckAt          time.Time `json:"last_check_at" gorm:"default:CURRENT_TIMESTAMP"`
	ErrorMessage         string    `json:"error_message" gorm:"type:text"`
	ConsecutiveFailures  int       `json:"consecutive_failures" gorm:"default:0"`
	TotalRequests        int       `json:"total_requests" gorm:"default:0"`
	SuccessfulRequests   int       `json:"successful_requests" gorm:"default:0"`
	FailedRequests       int       `json:"failed_requests" gorm:"default:0"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (ph *ProviderHealth) BeforeCreate(tx *gorm.DB) error {
	// Ensure unique constraint on provider_name and endpoint_url
	var count int64
	err := tx.Model(&ProviderHealth{}).
		Where("provider_name = ? AND endpoint_url = ?", ph.ProviderName, ph.EndpointURL).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("provider health record already exists for this provider and endpoint")
	}
	return nil
}

func CreateProviderHealth(ph *ProviderHealth) error {
	return DB.Create(ph).Error
}

func GetProviderHealth(providerName, endpointURL string) (*ProviderHealth, error) {
	if providerName == "" || endpointURL == "" {
		return nil, errors.New("provider name and endpoint URL are required")
	}

	var ph ProviderHealth
	err := DB.Where("provider_name = ? AND endpoint_url = ?", providerName, endpointURL).
		First(&ph).Error

	return &ph, HandleNotFound(err, ErrProviderHealthNotFound)
}

func UpdateProviderHealth(providerName, endpointURL string, updates map[string]interface{}) error {
	result := DB.Model(&ProviderHealth{}).
		Where("provider_name = ? AND endpoint_url = ?", providerName, endpointURL).
		Updates(updates)

	return HandleUpdateResult(result, ErrProviderHealthNotFound)
}

func GetProviderHealthByStatus(status string) ([]*ProviderHealth, error) {
	var providers []*ProviderHealth
	err := DB.Where("status = ?", status).Find(&providers).Error
	return providers, err
}

func GetAllProviderHealth(page, perPage int) ([]*ProviderHealth, int64, error) {
	var providers []*ProviderHealth
	var total int64

	tx := DB.Model(&ProviderHealth{})
	
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("last_check_at desc").Limit(limit).Offset(offset).Find(&providers).Error

	return providers, total, err
}

func SearchProviderHealth(keyword string, providerType, status string, page, perPage int) ([]*ProviderHealth, int64, error) {
	tx := DB.Model(&ProviderHealth{})

	if providerType != "" {
		tx = tx.Where("provider_type = ?", providerType)
	}

	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	if keyword != "" {
		tx = tx.Where("provider_name LIKE ? OR endpoint_url LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	var providers []*ProviderHealth
	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order("last_check_at desc").Limit(limit).Offset(offset).Find(&providers).Error

	return providers, total, err
}

func DeleteProviderHealth(id int) error {
	if id == 0 {
		return errors.New("provider health id is empty")
	}

	result := DB.Delete(&ProviderHealth{ID: id})
	return HandleUpdateResult(result, ErrProviderHealthNotFound)
}

func UpdateProviderHealthStats(providerName, endpointURL string, success bool, responseTimeMs int, errorMsg string) error {
	updates := map[string]interface{}{
		"last_check_at":     time.Now(),
		"response_time_ms":  responseTimeMs,
		"total_requests":    gorm.Expr("total_requests + ?", 1),
		"updated_at":        time.Now(),
	}

	if success {
		updates["successful_requests"] = gorm.Expr("successful_requests + ?", 1)
		updates["consecutive_failures"] = 0
		updates["status"] = ProviderStatusHealthy
		updates["error_message"] = ""
	} else {
		updates["failed_requests"] = gorm.Expr("failed_requests + ?", 1)
		updates["consecutive_failures"] = gorm.Expr("consecutive_failures + ?", 1)
		updates["error_message"] = errorMsg

		// Determine status based on consecutive failures
		var ph ProviderHealth
		err := DB.Where("provider_name = ? AND endpoint_url = ?", providerName, endpointURL).
			First(&ph).Error
		if err == nil {
			if ph.ConsecutiveFailures >= 5 {
				updates["status"] = ProviderStatusDown
			} else if ph.ConsecutiveFailures >= 2 {
				updates["status"] = ProviderStatusDegraded
			}
		}
	}

	// Calculate success rate
	var ph ProviderHealth
	err := DB.Where("provider_name = ? AND endpoint_url = ?", providerName, endpointURL).
		First(&ph).Error
	if err == nil {
		totalRequests := ph.TotalRequests + 1
		successfulRequests := ph.SuccessfulRequests
		if success {
			successfulRequests++
		}
		if totalRequests > 0 {
			successRate := float64(successfulRequests) / float64(totalRequests) * 100
			updates["success_rate"] = successRate
		}
	}

	result := DB.Model(&ProviderHealth{}).
		Where("provider_name = ? AND endpoint_url = ?", providerName, endpointURL).
		Updates(updates)

	return HandleUpdateResult(result, ErrProviderHealthNotFound)
}