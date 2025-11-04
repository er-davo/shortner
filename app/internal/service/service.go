package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"shortner/internal/models"
	"shortner/internal/repository"
)

type URLShortenerService struct {
	urlRepo    URLRepository
	clicksRepo ClicksRepository
}

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	GetByID(ctx context.Context, id int64) (*models.URL, error)
	GetByURL(ctx context.Context, shortened string) (*models.URL, error)
	Delete(ctx context.Context, id int64) error
}

type ClicksRepository interface {
	CreateClick(ctx context.Context, click *models.Click) error
	GetAnalitics(ctx context.Context, params *models.AnalyticsParams) (*models.AnalyticsResult, error)
}

func NewURLShortnererService(urlRepo URLRepository, clicksRepo ClicksRepository) *URLShortenerService {
	return &URLShortenerService{
		urlRepo:    urlRepo,
		clicksRepo: clicksRepo,
	}
}

func (s *URLShortenerService) Create(ctx context.Context, url *models.URL) error {
	var err error
	url.Shortened, err = s.generateShortenedURL(ctx, url.Original)
	if err != nil {
		return err
	}
	return s.urlRepo.Create(ctx, url)
}

func (s *URLShortenerService) GetByID(ctx context.Context, id int64) (*models.URL, error) {
	return s.urlRepo.GetByID(ctx, id)
}

func (s *URLShortenerService) GetByURL(ctx context.Context, shortened string) (*models.URL, error) {
	return s.urlRepo.GetByURL(ctx, shortened)
}

func (s *URLShortenerService) Delete(ctx context.Context, id int64) error {
	return s.urlRepo.Delete(ctx, id)
}

func (s *URLShortenerService) CreateClick(ctx context.Context, click *models.Click) error {
	return s.clicksRepo.CreateClick(ctx, click)
}

func (s *URLShortenerService) GetAnalytics(ctx context.Context, params *models.AnalyticsParams) (*models.AnalyticsResult, error) {
	return s.clicksRepo.GetAnalitics(ctx, params)
}

func (s *URLShortenerService) generateShortenedURL(ctx context.Context, original string) (string, error) {
	const (
		startingLength = 8
		maxLength      = 12
	)
	hash := md5.Sum([]byte(original))
	code := hex.EncodeToString(hash[:])

	for i := startingLength; i <= maxLength; i++ {
		shortened := code[:i]
		_, err := s.GetByURL(context.Background(), shortened)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return shortened, nil
			}
			return "", err
		}
	}

	return "", fmt.Errorf("failed to generate unique shortened url")
}
