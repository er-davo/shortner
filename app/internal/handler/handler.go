package handler

import (
	"net/http"
	"strconv"
	"time"

	"shortner/internal/models"
	"shortner/internal/service"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type URLHandler struct {
	urlService *service.URLShortenerService
	log        *zlog.Zerolog
}

func NewURLHandler(urlService *service.URLShortenerService, log *zlog.Zerolog) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		log:        log,
	}
}

// POST /urls
func (h *URLHandler) CreateShortenedURL(c *ginext.Context) {
	var req models.URL
	if err := c.BindJSON(&req); err != nil {
		h.log.Warn().Err(err).Msg("invalid request payload")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid request"})
		return
	}

	req.CreatedAt = time.Now()

	if err := h.urlService.Create(c, &req); err != nil {
		h.log.Error().Err(err).Msg("failed to create URL")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to create url"})
		return
	}

	h.log.Info().
		Str("original", req.Original).
		Str("shortened", req.Shortened).
		Msg("URL created successfully")

	c.JSON(http.StatusCreated, req)
}

// GET /:shortened
func (h *URLHandler) Redirect(c *ginext.Context) {
	shortened := c.Param("shortened")

	url, err := h.urlService.GetByURL(c, shortened)
	if err != nil {
		h.log.Warn().Str("shortened", shortened).Msg("url not found")
		c.JSON(http.StatusNotFound, ginext.H{"error": "url not found"})
		return
	}

	click := &models.Click{
		URLID:     url.ID,
		ClickedAt: time.Now(),
		UserAgent: c.Request.UserAgent(),
	}

	if err := h.urlService.CreateClick(c, click); err != nil {
		h.log.Warn().Err(err).Msg("failed to record click")
	}

	h.log.Info().
		Int64("url_id", url.ID).
		Str("destination", url.Original).
		Msg("redirecting user")

	c.Redirect(http.StatusFound, url.Original)
}

// DELETE /urls/:id
func (h *URLHandler) DeleteURL(c *ginext.Context) {
	strID := c.Param("id")
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		h.log.Warn().Str("id", strID).Msg("invalid id format")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid id"})
		return
	}

	if err := h.urlService.Delete(c, id); err != nil {
		h.log.Error().Err(err).Int64("id", id).Msg("failed to delete URL")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to delete url"})
		return
	}

	h.log.Info().Int64("id", id).Msg("URL deleted successfully")
	c.JSON(http.StatusOK, ginext.H{"deleted": id})
}

// POST /analytics
func (h *URLHandler) GetAnalytics(c *ginext.Context) {
	var params models.AnalyticsParams
	if err := c.BindJSON(&params); err != nil {
		h.log.Warn().Err(err).Msg("invalid analytics params")
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid params"})
		return
	}

	res, err := h.urlService.GetAnalytics(c, &params)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get analytics")
		c.JSON(http.StatusInternalServerError, ginext.H{"error": "failed to get analytics"})
		return
	}

	h.log.Info().
		Str("url", params.URL).
		Int("total_clicks", res.TotalClicks).
		Msg("analytics fetched successfully")

	c.JSON(http.StatusOK, res)
}

// Register all routes
func (h *URLHandler) RegisterRoutes(r *ginext.Engine) {
	r.POST("/urls", h.CreateShortenedURL)
	r.GET("/:shortened", h.Redirect)
	r.DELETE("/urls/:id", h.DeleteURL)
	r.POST("/analytics", h.GetAnalytics)
}
