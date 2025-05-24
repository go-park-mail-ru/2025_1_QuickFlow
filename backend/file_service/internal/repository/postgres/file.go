package postgres

import (
	"context"
	"database/sql"
	"errors"

	"quickflow/shared/models"
)

const (
	addFileQuery = `
        INSERT INTO files (file_url, filename)
        VALUES ($1, $2)
    `
)

type PostgresFileRepository struct {
	connPool *sql.DB
}

// NewPostgresFileRepository создает новый экземпляр репозитория для работы с файлами
func NewPostgresFileRepository(db *sql.DB) *PostgresFileRepository {
	return &PostgresFileRepository{connPool: db}
}

// Close закрывает пул соединений
func (p *PostgresFileRepository) Close() {
	if p.connPool != nil {
		p.connPool.Close()
	}
}

func (p *PostgresFileRepository) AddFileRecord(ctx context.Context, file *models.File) error {
	if file == nil {
		return errors.New("file cannot be nil")
	}
	_, err := p.connPool.ExecContext(ctx, addFileQuery, file.URL, file.Name)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresFileRepository) AddFilesRecords(ctx context.Context, files []*models.File) error {
	tx, err := p.connPool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, file := range files {
		if file == nil {
			return errors.New("file cannot be nil")
		}
		_, err := tx.ExecContext(ctx, addFileQuery, file.URL, file.Name)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
