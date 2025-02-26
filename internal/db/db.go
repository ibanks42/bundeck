package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Schema version table to track migrations
const schemaVersionTable = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER PRIMARY KEY,
	applied_at DATETIME NOT NULL
);`

// All migrations in order
var migrations = []string{
	// v1: Initial schema
	`CREATE TABLE IF NOT EXISTS plugins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		order_num INTEGER NOT NULL,
		image BLOB,
		image_type TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`,
	// v2/3: Add continuous running support
	`ALTER TABLE plugins ADD COLUMN run_continuously BOOLEAN NOT NULL DEFAULT 0;`,
	`ALTER TABLE plugins ADD COLUMN interval_seconds INTEGER NOT NULL DEFAULT 0;`,
}

func getCurrentVersion(db *sql.DB) (int, error) {
	// Create version table if it doesn't exist
	if _, err := db.Exec(schemaVersionTable); err != nil {
		return 0, fmt.Errorf("failed to create schema version table: %w", err)
	}

	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current schema version: %w", err)
	}

	return version, nil
}

func InitDB(db *sql.DB) error {
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return err
	}

	// Apply all migrations that haven't been applied yet
	for version := currentVersion + 1; version <= len(migrations); version++ {
		migration := migrations[version-1]

		// Start transaction for each migration
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction for version %d: %w", version, err)
		}

		// Apply migration
		if _, err := tx.Exec(migration); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration version %d: %w", version, err)
		}

		// Record migration version
		if _, err := tx.Exec(
			"INSERT INTO schema_version (version, applied_at) VALUES (?, ?)",
			version,
			time.Now(),
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration version %d: %w", version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration version %d: %w", version, err)
		}

		fmt.Printf("Applied migration version %d\n", version)
	}

	// Do periodic WAL checkpoints
	checkpoint(db)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			checkpoint(db)
		}
	}()

	return nil
}

func checkpoint(db *sql.DB) {
	_, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	if err != nil {
		_ = fmt.Errorf("failed to truncate WAL file: %w", err)
	}
}

type Plugin struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Code            string    `json:"code"`
	OrderNum        int       `json:"order_num"`
	Image           []byte    `json:"image"`
	ImageType       *string   `json:"image_type"`
	RunContinuously bool      `json:"run_continuously"`
	IntervalSeconds int       `json:"interval_seconds"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PluginStore struct {
	db *sql.DB
}

func NewPluginStore(db *sql.DB) *PluginStore {
	return &PluginStore{db: db}
}

func (s *PluginStore) Create(plugin *Plugin) error {
	now := time.Now()
	plugin.CreatedAt = now
	plugin.UpdatedAt = now

	result, err := s.db.Exec(
		"INSERT INTO plugins (name, code, order_num, image, image_type, run_continuously, interval_seconds, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		plugin.Name,
		plugin.Code,
		plugin.OrderNum,
		plugin.Image,
		plugin.ImageType,
		plugin.RunContinuously,
		plugin.IntervalSeconds,
		plugin.CreatedAt,
		plugin.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	plugin.ID = int(id)
	return nil
}

func (s *PluginStore) GetAll() ([]Plugin, error) {
	rows, err := s.db.Query("SELECT id, name, code, order_num, image, image_type, run_continuously, interval_seconds, created_at, updated_at FROM plugins ORDER BY order_num")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plugins []Plugin
	for rows.Next() {
		var p Plugin
		var imageType sql.NullString // Use sql.NullString for nullable column
		err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.OrderNum, &p.Image, &imageType, &p.RunContinuously, &p.IntervalSeconds, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if imageType.Valid {
			p.ImageType = &imageType.String
		}
		plugins = append(plugins, p)
	}

	return plugins, nil
}

func (s *PluginStore) GetByID(id int) (*Plugin, error) {
	var p Plugin
	var imageType sql.NullString // Use sql.NullString for nullable column
	err := s.db.QueryRow(
		"SELECT id, name, code, order_num, image, image_type, run_continuously, interval_seconds, created_at, updated_at FROM plugins WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.Code, &p.OrderNum, &p.Image, &imageType, &p.RunContinuously, &p.IntervalSeconds, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if imageType.Valid {
		p.ImageType = &imageType.String
	}
	return &p, nil
}

func (s *PluginStore) UpdateCode(id int, code string, image []byte, imageType string, name string, runContinuously bool, intervalSeconds int) error {
	result, err := s.db.Exec(
		"UPDATE plugins SET code = ?, image = ?, image_type = ?, name = ?, run_continuously = ?, interval_seconds = ?, updated_at = ? WHERE id = ?",
		code,
		image,
		imageType,
		name,
		runContinuously,
		intervalSeconds,
		time.Now(),
		id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *PluginStore) UpdateOrder(orders []struct {
	ID       int `json:"id"`
	OrderNum int `json:"order_num"`
},
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, o := range orders {
		_, err := tx.Exec(
			"UPDATE plugins SET order_num = ?, updated_at = ? WHERE id = ?",
			o.OrderNum,
			time.Now(),
			o.ID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *PluginStore) Delete(id int) error {
	result, err := s.db.Exec("DELETE FROM plugins WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
