package cooks

import (
	"context"

	"github.com/superryanguo/kitchen/cooks/models"
)

type Repository interface {
	GetCookByID(ctx context.Context, id int) *models.Cook
	GetAvailableCooks(ctx context.Context) *[]models.Cook
	GetFirstAvailableCook(ctx context.Context, cookCh chan *models.Cook)
	UpdateCookStatus(ctx context.Context, cookID int, status int) error
}
