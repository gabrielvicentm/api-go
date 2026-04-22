package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gabrielvicentm/api-go.git/internal/domain"
	"github.com/gin-gonic/gin"
)

func respondSuccess(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{
		"message": message,
		"data":    data,
	})
}

func respondList(c *gin.Context, message string, data any, page, limit int, total int64) {
	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"data":    data,
		"meta": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func respondDomainError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": fallbackMessage})
	}
}

func parsePagination(c *gin.Context) (int, int) {
	page := 1
	limit := 20

	if rawPage := c.Query("page"); rawPage != "" {
		if parsed, err := strconv.Atoi(rawPage); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if rawLimit := c.Query("limit"); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 {
			if parsed > 100 {
				parsed = 100
			}
			limit = parsed
		}
	}

	return page, limit
}
