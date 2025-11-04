package repository

import (
	"context"

	"shortner/internal/models"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type ClicksRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewClicksRepository(db *dbpg.DB, strategy retry.Strategy) *ClicksRepository {
	return &ClicksRepository{
		db:       db,
		strategy: strategy,
	}
}

func (r *ClicksRepository) CreateClick(ctx context.Context, click *models.Click) error {
	if click == nil {
		return ErrNilValue
	}

	query := `
		INSERT INTO clicks (url_id, clicked_at, user_agent)
		VALUES ($1, $2, $3) RETURNING id
	`

	row, err := r.db.QueryRowWithRetry(
		ctx,
		r.strategy,
		query,
		click.URLID,
		click.ClickedAt,
		click.UserAgent,
	)
	if err != nil {
		return wrapDBError(err)
	}

	err = row.Scan(&click.ID)

	return wrapDBError(err)
}

func (r *ClicksRepository) GetAnalitics(ctx context.Context, params *models.AnalyticsParams) (*models.AnalyticsResult, error) {
	if params.URL == "" {
		return nil, ErrInvalidValue
	}

	var urlID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM urls WHERE shortened = $1`,
		params.URL,
	).Scan(&urlID)
	if err != nil {
		return nil, wrapDBError(err)
	}

	var groupExpr string
	switch params.GroupBy {
	case models.ByDay:
		groupExpr = "DATE(clicked_at)"
	case models.ByMonth:
		groupExpr = "TO_CHAR(clicked_at, 'YYYY-MM')"
	case models.ByYear:
		groupExpr = "TO_CHAR(clicked_at, 'YYYY')"
	case models.ByUserAgent:
		groupExpr = "user_agent"
	default:
		groupExpr = "DATE(clicked_at)"
	}

	query := `
		SELECT
			` + groupExpr + ` AS grp,
			COUNT(*) AS cnt
		FROM clicks
		WHERE url_id = $1
			AND clicked_at BETWEEN $2 AND $3
		GROUP BY grp
		ORDER BY grp;
	`

	rows, err := r.db.QueryContext(ctx, query, urlID, params.From, params.To)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	result := &models.AnalyticsResult{
		Grouped: make(map[string]int),
	}

	for rows.Next() {
		var grp string
		var cnt int
		if err := rows.Scan(&grp, &cnt); err != nil {
			return nil, wrapDBError(err)
		}
		result.Grouped[grp] = cnt
		result.TotalClicks += cnt
	}

	if err := rows.Err(); err != nil {
		return nil, wrapDBError(err)
	}

	return result, nil
}
