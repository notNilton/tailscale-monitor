package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nilbyte-studios/network-infra/internal/metrics"
)

// Storage gerencia o armazenamento de métricas em SQLite
type Storage struct {
	db *sql.DB
}

// NewStorage cria uma nova instância de Storage
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}

	if err := storage.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return storage, nil
}

// Close fecha a conexão com o banco de dados
func (s *Storage) Close() error {
	return s.db.Close()
}

// migrate cria as tabelas necessárias
func (s *Storage) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		hostname TEXT NOT NULL,
		data TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_timestamp ON metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_hostname ON metrics(hostname);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Save salva métricas no banco de dados
func (s *Storage) Save(m *metrics.SystemMetrics) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	query := `INSERT INTO metrics (timestamp, hostname, data) VALUES (?, ?, ?)`
	_, err = s.db.Exec(query, m.Timestamp, m.Hostname, string(data))
	if err != nil {
		return fmt.Errorf("failed to insert metrics: %w", err)
	}

	return nil
}

// GetLatest retorna as métricas mais recentes
func (s *Storage) GetLatest() (*metrics.SystemMetrics, error) {
	query := `SELECT data FROM metrics ORDER BY timestamp DESC LIMIT 1`

	var data string
	err := s.db.QueryRow(query).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest metrics: %w", err)
	}

	var m metrics.SystemMetrics
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	return &m, nil
}

// GetHistory retorna histórico de métricas
func (s *Storage) GetHistory(hours int) ([]metrics.SystemMetrics, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	query := `SELECT data FROM metrics WHERE timestamp >= ? ORDER BY timestamp DESC`

	rows, err := s.db.Query(query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var history []metrics.SystemMetrics
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			continue
		}

		var m metrics.SystemMetrics
		if err := json.Unmarshal([]byte(data), &m); err != nil {
			continue
		}

		history = append(history, m)
	}

	return history, nil
}

// Cleanup remove métricas antigas
func (s *Storage) Cleanup(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	query := `DELETE FROM metrics WHERE timestamp < ?`
	result, err := s.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old metrics: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		fmt.Printf("Cleaned up %d old metric records\n", rows)
	}

	return nil
}
