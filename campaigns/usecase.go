package campaigns

import (
	"github.com/aristio/count/models"

	"github.com/labstack/echo"
)

// UseCase represent the campaign's usecases
type UseCase interface {
	CreateCampaign(echo.Context, *models.Campaign) error
	UpdateCampaign(echo.Context, string, *models.Campaign) error
	GetCampaignDetail(echo.Context, string) (*models.Campaign, error)
	GetCampaign(echo.Context, map[string]interface{}) (string, []*models.Campaign, error)
	UpdateStatusBasedOnStartDate() error
	GetCampaignAvailable(echo.Context, string) ([]*models.Campaign, error)
}
