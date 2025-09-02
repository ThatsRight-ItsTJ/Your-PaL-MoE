package model

import (
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Version     string `json:"version" gorm:"uniqueIndex"`
	Description string `json:"description"`
	Applied     bool   `json:"applied"`
}

// TableName returns the table name for Migration
func (Migration) TableName() string {
	return "migrations"
}

// MigrationManager handles database migrations
type MigrationManager struct {
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager() *MigrationManager {
	return &MigrationManager{
		migrations: []Migration{
			{Version: "001", Description: "Create initial tables", Applied: false},
			{Version: "002", Description: "Add audit log table", Applied: false},
			{Version: "003", Description: "Add user management", Applied: false},
		},
	}
}

// GetPendingMigrations returns migrations that haven't been applied
func (m *MigrationManager) GetPendingMigrations() []Migration {
	var pending []Migration
	for _, migration := range m.migrations {
		if !migration.Applied {
			pending = append(pending, migration)
		}
	}
	return pending
}

// ApplyMigration marks a migration as applied
func (m *MigrationManager) ApplyMigration(version string) error {
	for i, migration := range m.migrations {
		if migration.Version == version {
			m.migrations[i].Applied = true
			log.Printf("Applied migration %s: %s", version, migration.Description)
			return nil
		}
	}
	return fmt.Errorf("migration %s not found", version)
}

// GetAllMigrations returns all migrations
func (m *MigrationManager) GetAllMigrations() []Migration {
	return m.migrations
}