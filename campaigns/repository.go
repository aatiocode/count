package campaigns

import (
	"github.com/aristio/count/models"

	"github.com/labstack/echo"
)

// Repository represent the campaign's repository contract
type Repository interface {
	CreateCampaign(echo.Context, *models.Campaign) error
	UpdateCampaign(echo.Context, int64, *models.Campaign) error
	GetCampaignDetail(echo.Context, int64) (*models.Campaign, error)
	GetCampaign(echo.Context, map[string]interface{}) ([]*models.Campaign, error)
	SavePoint(echo.Context, *models.CampaignTrx) error
	CountCampaign(echo.Context, map[string]interface{}) (int, error)
	UpdateExpiryDate(echo.Context) error
	UpdateStatusBasedOnStartDate() error
	GetCampaignAvailable(echo.Context, string) ([]*models.Campaign, error)
	Delete(echo.Context, int64) error
}
