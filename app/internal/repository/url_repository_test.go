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

func TestURLRepository_CRUD(t *testing.T) {
	strategy := retry.Strategy{
		Attempts: 1,
		Delay:    1 * time.Second,
		Backoff:  1,
	}
	repo := repository.NewURLRepository(db, strategy)

	url := &models.URL{
		Original:  "https://example.com/test",
		Shortened: "abc123",
		CreatedAt: time.Now(),
	}

	t.Run("Create", func(t *testing.T) {
		err := repo.Create(t.Context(), url)
		assert.NoError(t, err)
		assert.NotZero(t, url.ID)
	})

	t.Run("GetByID", func(t *testing.T) {
		got, err := repo.GetByID(t.Context(), url.ID)
		assert.NoError(t, err)
		assert.Equal(t, url.Original, got.Original)
		assert.Equal(t, url.Shortened, got.Shortened)
	})

	t.Run("GetByURL", func(t *testing.T) {
		got, err := repo.GetByURL(t.Context(), url.Shortened)
		assert.NoError(t, err)
		assert.Equal(t, url.ID, got.ID)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(t.Context(), url.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(t.Context(), url.ID)
		assert.Error(t, err)
	})
}
