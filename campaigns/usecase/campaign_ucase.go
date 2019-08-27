package usecase

import (
	"errors"
	"github.com/aristio/count/campaigns"
	"github.com/aristio/count/models"
	"github.com/aristio/count/rewards"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

type campaignUseCase struct {
	campaignRepo campaigns.Repository
	rewardUC     rewards.UseCase
}

// NewCampaignUseCase will create new an campaignUseCase object representation of campaigns.UseCase interface
func NewCampaignUseCase(cmpgn campaigns.Repository, rwd rewards.UseCase) campaigns.UseCase {
	return &campaignUseCase{
		campaignRepo: cmpgn,
		rewardUC:     rwd,
	}
}

func (cmpgn *campaignUseCase) CreateCampaign(c echo.Context, campaign *models.Campaign) error {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	err := cmpgn.campaignRepo.CreateCampaign(c, campaign)

	if err != nil {
		requestLogger.Debug(models.ErrCampaignFailed)

		return models.ErrCampaignFailed
	}

	// create array rewards
	for _, reward := range *campaign.Rewards {
		err = cmpgn.rewardUC.CreateReward(c, &reward, campaign)

		if err != nil {
			break
		}
	}

	if err != nil {
		_ = cmpgn.rewardUC.DeleteByCampaign(c, campaign.ID)
		_ = cmpgn.campaignRepo.Delete(c, campaign.ID)

		requestLogger.Debug(models.ErrCampaignFailed)

		return models.ErrCreateRewardsFailed
	}

	return nil
}

func (cmpgn *campaignUseCase) UpdateCampaign(c echo.Context, id string, updateCampaign *models.Campaign) error {
	var campaignDetail *models.Campaign
	now := time.Now()
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	campaignID, err := strconv.Atoi(id)

	if err != nil {
		requestLogger.Debug(err)

		return errors.New("Something went wrong with input ID")
	}

	campaignDetail, err = cmpgn.campaignRepo.GetCampaignDetail(c, int64(campaignID))

	if err != nil {
		requestLogger.Debug(models.ErrNoCampaign)

		return models.ErrNoCampaign
	}

	vEndDate, _ := time.Parse(time.RFC3339, campaignDetail.EndDate)

	if vEndDate.Before(now.Add(time.Hour * -24)) {
		requestLogger.Debug(models.ErrCampaignExpired)

		return models.ErrCampaignExpired
	}

	err = cmpgn.campaignRepo.UpdateCampaign(c, int64(campaignID), updateCampaign)

	if err != nil {
		requestLogger.Debug(models.ErrCampaignUpdateFailed)

		return models.ErrCampaignUpdateFailed
	}

	return nil
}

func (cmpgn *campaignUseCase) GetCampaign(c echo.Context, payload map[string]interface{}) (string, []*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	page, err := strconv.Atoi(payload["page"].(string))

	if err != nil {
		requestLogger.Debug(err)

		return "", nil, errors.New("Something went wrong with input page")
	}

	limit, err := strconv.Atoi(payload["limit"].(string))

	if err != nil {
		requestLogger.Debug(err)

		return "", nil, errors.New("Something went wrong with input limit")
	}

	payload["page"] = page
	payload["limit"] = limit
	_ = cmpgn.campaignRepo.UpdateExpiryDate(c)
	listCampaign, err := cmpgn.campaignRepo.GetCampaign(c, payload)

	if err != nil {
		requestLogger.Debug(models.ErrGetCampaign)

		return "", nil, models.ErrGetCampaign
	}

	countCampaign, err := cmpgn.campaignRepo.CountCampaign(c, payload)

	if err != nil {
		requestLogger.Debug(models.ErrGetCampaign)

		return "", nil, err
	}

	if countCampaign <= 0 {
		requestLogger.Debug(models.ErrGetCampaignCounter)

		return "", listCampaign, models.ErrGetCampaignCounter
	}

	return strconv.Itoa(countCampaign), listCampaign, nil
}

func (cmpgn *campaignUseCase) GetCampaignDetail(c echo.Context, id string) (*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	campaignID, err := strconv.Atoi(id)

	if err != nil {
		requestLogger.Debug(err)

		return nil, errors.New("Something went wrong with input ID")
	}

	campaignDetail, err := cmpgn.campaignRepo.GetCampaignDetail(c, int64(campaignID))

	if err != nil {
		requestLogger.Debug(models.ErrNoCampaign)

		return nil, models.ErrNoCampaign
	}

	return campaignDetail, nil
}

func (cmpgn *campaignUseCase) GetCampaignAvailable(c echo.Context, today string) ([]*models.Campaign, error) {
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	campaigns, err := cmpgn.campaignRepo.GetCampaignAvailable(c, today)

	if err != nil {
		requestLogger.Debug(models.ErrNoCampaign)

		return nil, models.ErrNoCampaign
	}

	return campaigns, nil
}

func (cmpgn *campaignUseCase) UpdateStatusBasedOnStartDate() error {
	err := cmpgn.campaignRepo.UpdateStatusBasedOnStartDate()

	if err != nil {
		logrus.Debug("Update Status Base on Start Date: ", err)

		return err
	}
	return nil
}
