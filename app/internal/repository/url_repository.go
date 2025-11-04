package repository

import (
	"context"
	"shortner/internal/models"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type URLRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewURLRepository(db *dbpg.DB, strategy retry.Strategy) *URLRepository {
	return &URLRepository{
		db:       db,
		strategy: strategy,
	}
}

func (r *URLRepository) Create(ctx context.Context, url *models.URL) error {
	if url == nil {
		return ErrNilValue
	}
	if url.Original == "" || url.Shortened == "" {
		return ErrInvalidValue
	}

	query := `
		INSERT INTO urls (original, shortened, created_at)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	row, err := r.db.QueryRowWithRetry(
		ctx,
		r.strategy,
		query,
		url.Original,
		url.Shortened,
		url.CreatedAt,
	)
	if err != nil {
		return wrapDBError(err)
	}

	err = row.Scan(&url.ID)
	return wrapDBError(err)
}

func (r *URLRepository) GetByID(ctx context.Context, id int64) (*models.URL, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	query := `
		SELECT id, original, shortened, created_at
		FROM urls
		WHERE id = $1;
	`

	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, id)
	if err != nil {
		return nil, wrapDBError(err)
	}

	url := &models.URL{}
	err = row.Scan(&url.ID, &url.Original, &url.Shortened, &url.CreatedAt)
	return url, wrapDBError(err)
}

func (r *URLRepository) GetByURL(ctx context.Context, shortened string) (*models.URL, error) {
	if shortened == "" {
		return nil, ErrInvalidValue
	}

	query := `
		SELECT id, original, shortened, created_at
		FROM urls
		WHERE shortened = $1;
	`

	row, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, shortened)
	if err != nil {
		return nil, wrapDBError(err)
	}

	url := &models.URL{}
	err = row.Scan(&url.ID, &url.Original, &url.Shortened, &url.CreatedAt)
	return url, wrapDBError(err)
}

func (r *URLRepository) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidID
	}

	query := `DELETE FROM urls WHERE id = $1;`

	_, err := r.db.ExecWithRetry(ctx, r.strategy, query, id)
	return wrapDBError(err)
}
