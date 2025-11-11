//go:build integration
// +build integration

package repository_test

import (
	"shortner/internal/models"
	"shortner/internal/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wb-go/wbf/retry"
)

func TestClicksRepository_CreateAndAnalytics(t *testing.T) {
	strategy := retry.Strategy{
		Attempts: 1,
		Delay:    1 * time.Second,
		Backoff:  1,
	}
	urlRepo := repository.NewURLRepository(db, strategy)
	clickRepo := repository.NewClicksRepository(db, strategy)

	url := &models.URL{
		Original:  "https://example.com/page",
		Shortened: "xyz123",
		CreatedAt: time.Now(),
	}
	assert.NoError(t, urlRepo.Create(t.Context(), url))

	now := time.Now()

	clicks := []*models.Click{
		{URLID: url.ID, ClickedAt: now.Add(-48 * time.Hour), UserAgent: "Chrome"},
		{URLID: url.ID, ClickedAt: now.Add(-24 * time.Hour), UserAgent: "Firefox"},
		{URLID: url.ID, ClickedAt: now.Add(-24 * time.Hour), UserAgent: "Chrome"},
		{URLID: url.ID, ClickedAt: now, UserAgent: "Chrome"},
	}

	for _, c := range clicks {
		err := clickRepo.CreateClick(t.Context(), c)
		assert.NoError(t, err)
		assert.NotZero(t, c.ID)
	}

	t.Run("Group by day", func(t *testing.T) {
		params := &models.AnalyticsParams{
			URL:     url.Shortened,
			From:    now.Add(-72 * time.Hour),
			To:      now.Add(24 * time.Hour),
			GroupBy: models.ByDay,
		}

		res, err := clickRepo.GetAnalitics(t.Context(), params)
		assert.NoError(t, err)
		assert.Equal(t, len(clicks), res.TotalClicks)
		assert.GreaterOrEqual(t, len(res.Grouped), 2)
	})

	t.Run("Group by user agent", func(t *testing.T) {
		params := &models.AnalyticsParams{
			URL:     url.Shortened,
			From:    now.Add(-72 * time.Hour),
			To:      now.Add(24 * time.Hour),
			GroupBy: models.ByUserAgent,
		}

		res, err := clickRepo.GetAnalitics(t.Context(), params)
		assert.NoError(t, err)
		assert.Equal(t, len(clicks), res.TotalClicks)
		assert.Equal(t, 2, len(res.Grouped))
		assert.Equal(t, 3, res.Grouped["Chrome"])
		assert.Equal(t, 1, res.Grouped["Firefox"])
	})
}
