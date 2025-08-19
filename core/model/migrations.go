package model

import (
	"github.com/labring/aiproxy/core/common"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RunEnhancedMigrations runs additional migrations for enhanced features
func RunEnhancedMigrations() error {
	log.Info("Running enhanced migrations for BricksLLM-style features")

	// Add enhanced fields to existing tokens table
	if err := addTokenEnhancedFields(); err != nil {
		return err
	}

	// Create indexes for better performance
	if err := createEnhancedIndexes(); err != nil {
		return err
	}

	log.Info("Enhanced migrations completed successfully")
	return nil
}

func addTokenEnhancedFields() error {
	// Check if we're using a database that supports ALTER TABLE
	if common.UsingSQLite {
		return addTokenEnhancedFieldsSQLite()
	} else if common.UsingPostgreSQL {
		return addTokenEnhancedFieldsPostgreSQL()
	} else if common.UsingMySQL {
		return addTokenEnhancedFieldsMySQL()
	}

	return nil
}

func addTokenEnhancedFieldsSQLite() error {
	alterStatements := []string{
		// Cost tracking and limits
		"ALTER TABLE tokens ADD COLUMN cost_limit_usd DECIMAL(10,4) DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN cost_used_usd DECIMAL(10,4) DEFAULT 0.00",
		"ALTER TABLE tokens ADD COLUMN cost_reset_date DATETIME DEFAULT NULL",

		// Usage quotas and rate limits  
		"ALTER TABLE tokens ADD COLUMN rate_limit_rpm INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN rate_limit_rph INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN rate_limit_rpd INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN quota_requests_daily INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN quota_requests_monthly INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN requests_used_daily INTEGER DEFAULT 0",
		"ALTER TABLE tokens ADD COLUMN requests_used_monthly INTEGER DEFAULT 0",
		"ALTER TABLE tokens ADD COLUMN last_request_time DATETIME DEFAULT NULL",

		// Tag-based organization
		"ALTER TABLE tokens ADD COLUMN tags TEXT DEFAULT '[]'",
		"ALTER TABLE tokens ADD COLUMN description TEXT DEFAULT ''",
		"ALTER TABLE tokens ADD COLUMN environment VARCHAR(50) DEFAULT 'production'",

		// Key lifecycle management
		"ALTER TABLE tokens ADD COLUMN expires_at DATETIME DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN auto_rotate BOOLEAN DEFAULT FALSE",
		"ALTER TABLE tokens ADD COLUMN rotation_interval_days INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN last_rotated_at DATETIME DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN token_status VARCHAR(20) DEFAULT 'active'",

		// Model access control
		"ALTER TABLE tokens ADD COLUMN allowed_models TEXT DEFAULT '[]'",
		"ALTER TABLE tokens ADD COLUMN blocked_models TEXT DEFAULT '[]'",
		"ALTER TABLE tokens ADD COLUMN allowed_endpoints TEXT DEFAULT '[]'",

		// Enhanced metadata
		"ALTER TABLE tokens ADD COLUMN created_by INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN last_used_at DATETIME DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN usage_count INTEGER DEFAULT 0",
	}

	return executeAlterStatements(alterStatements)
}

func addTokenEnhancedFieldsPostgreSQL() error {
	alterStatements := []string{
		// Cost tracking and limits
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS cost_limit_usd DECIMAL(10,4) DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS cost_used_usd DECIMAL(10,4) DEFAULT 0.00",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS cost_reset_date TIMESTAMP DEFAULT NULL",

		// Usage quotas and rate limits  
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS rate_limit_rpm INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS rate_limit_rph INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS rate_limit_rpd INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS quota_requests_daily INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS quota_requests_monthly INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS requests_used_daily INTEGER DEFAULT 0",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS requests_used_monthly INTEGER DEFAULT 0",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS last_request_time TIMESTAMP DEFAULT NULL",

		// Tag-based organization
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '[]'",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS description TEXT DEFAULT ''",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS environment VARCHAR(50) DEFAULT 'production'",

		// Key lifecycle management
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS auto_rotate BOOLEAN DEFAULT FALSE",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS rotation_interval_days INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS last_rotated_at TIMESTAMP DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS token_status VARCHAR(20) DEFAULT 'active'",

		// Model access control
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS allowed_models JSONB DEFAULT '[]'",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS blocked_models JSONB DEFAULT '[]'",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS allowed_endpoints JSONB DEFAULT '[]'",

		// Enhanced metadata
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS created_by INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN IF NOT EXISTS usage_count INTEGER DEFAULT 0",
	}

	return executeAlterStatements(alterStatements)
}

func addTokenEnhancedFieldsMySQL() error {
	alterStatements := []string{
		// Cost tracking and limits
		"ALTER TABLE tokens ADD COLUMN cost_limit_usd DECIMAL(10,4) DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN cost_used_usd DECIMAL(10,4) DEFAULT 0.00",
		"ALTER TABLE tokens ADD COLUMN cost_reset_date TIMESTAMP DEFAULT NULL",

		// Usage quotas and rate limits  
		"ALTER TABLE tokens ADD COLUMN rate_limit_rpm INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN rate_limit_rph INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN rate_limit_rpd INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN quota_requests_daily INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN quota_requests_monthly INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN requests_used_daily INTEGER DEFAULT 0",
		"ALTER TABLE tokens ADD COLUMN requests_used_monthly INTEGER DEFAULT 0",
		"ALTER TABLE tokens ADD COLUMN last_request_time TIMESTAMP DEFAULT NULL",

		// Tag-based organization
		"ALTER TABLE tokens ADD COLUMN tags JSON DEFAULT (JSON_ARRAY())",
		"ALTER TABLE tokens ADD COLUMN description TEXT DEFAULT ''",
		"ALTER TABLE tokens ADD COLUMN environment VARCHAR(50) DEFAULT 'production'",

		// Key lifecycle management
		"ALTER TABLE tokens ADD COLUMN expires_at TIMESTAMP DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN auto_rotate BOOLEAN DEFAULT FALSE",
		"ALTER TABLE tokens ADD COLUMN rotation_interval_days INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN last_rotated_at TIMESTAMP DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN token_status VARCHAR(20) DEFAULT 'active'",

		// Model access control
		"ALTER TABLE tokens ADD COLUMN allowed_models JSON DEFAULT (JSON_ARRAY())",
		"ALTER TABLE tokens ADD COLUMN blocked_models JSON DEFAULT (JSON_ARRAY())",
		"ALTER TABLE tokens ADD COLUMN allowed_endpoints JSON DEFAULT (JSON_ARRAY())",

		// Enhanced metadata
		"ALTER TABLE tokens ADD COLUMN created_by INTEGER DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN last_used_at TIMESTAMP DEFAULT NULL",
		"ALTER TABLE tokens ADD COLUMN usage_count INTEGER DEFAULT 0",
	}

	return executeAlterStatements(alterStatements)
}

func executeAlterStatements(statements []string) error {
	for _, stmt := range statements {
		if err := DB.Exec(stmt).Error; err != nil {
			// Log warning but continue - column might already exist
			log.Warnf("Failed to execute migration statement '%s': %v", stmt, err)
		}
	}
	return nil
}

func createEnhancedIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_tokens_created_by ON tokens(created_by)",
		"CREATE INDEX IF NOT EXISTS idx_tokens_last_used_at ON tokens(last_used_at)",
		"CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at)",
		"CREATE INDEX IF NOT EXISTS idx_tokens_environment ON tokens(environment)",
		"CREATE INDEX IF NOT EXISTS idx_tokens_token_status ON tokens(token_status)",
		"CREATE INDEX IF NOT EXISTS idx_provider_health_status ON provider_health(status)",
		"CREATE INDEX IF NOT EXISTS idx_provider_health_provider ON provider_health(provider_name)",
		"CREATE INDEX IF NOT EXISTS idx_provider_health_type ON provider_health(provider_type)",
		"CREATE INDEX IF NOT EXISTS idx_task_executions_user ON task_executions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_task_executions_token ON task_executions(token_id)",
		"CREATE INDEX IF NOT EXISTS idx_task_executions_status ON task_executions(status)",
		"CREATE INDEX IF NOT EXISTS idx_subtask_executions_task ON subtask_executions(task_execution_id)",
		"CREATE INDEX IF NOT EXISTS idx_subtask_executions_provider ON subtask_executions(provider_name)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_token ON audit_logs(token_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at)",
	}

	for _, indexStmt := range indexes {
		if err := DB.Exec(indexStmt).Error; err != nil {
			log.Warnf("Failed to create index: %v", err)
		}
	}

	return nil
}