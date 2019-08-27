package http

import (
	"github.com/aristio/count/campaigns"
	"github.com/aristio/count/models"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

var response models.Response

// CampaignsHandler represent the httphandler for campaigns
type CampaignsHandler struct {
	CampaignUseCase campaigns.UseCase
}

// NewCampaignsHandler represent to register campaigns endpoint
func NewCampaignsHandler(echoGroup models.EchoGroup, us campaigns.UseCase) {
	handler := &CampaignsHandler{
		CampaignUseCase: us,
	}

	// End Point For CMS
	echoGroup.Admin.POST("/campaigns", handler.CreateCampaign)
	echoGroup.Admin.PUT("/campaigns/status/:id", handler.UpdateStatusCampaign)
	echoGroup.Admin.GET("/campaigns", handler.GetCampaigns)
	echoGroup.Admin.GET("/campaigns/:id", handler.GetCampaignDetail)
}

// CreateCampaign a handler to create a campaign
func (cmpgn *CampaignsHandler) CreateCampaign(c echo.Context) error {
	var campaign models.Campaign
	response = models.Response{}
	logger := models.RequestLogger{}
	err := c.Bind(&campaign)

	if err != nil {
		response.Status = models.StatusError
		response.Message = err.Error()
		return c.JSON(getStatusCode(err), response)
	}

	requestLogger := logger.GetRequestLogger(c, campaign)
	requestLogger.Info("Start to create a campaign.")
	err = cmpgn.CampaignUseCase.CreateCampaign(c, &campaign)

	if err != nil {
		response.Status = models.StatusError
		response.Message = err.Error()
		return c.JSON(getStatusCode(err), response)
	}

	if (models.Campaign{}) != campaign {
		response.Data = campaign
	}

	response.Status = models.StatusSuccess
	response.Message = models.MessageSaveSuccess

	requestLogger.Info("End of create a campaign.")

	return c.JSON(http.StatusCreated, response)
}

// UpdateStatusCampaign a handler to update campaign status
func (cmpgn *CampaignsHandler) UpdateStatusCampaign(c echo.Context) error {
	updateCampaign := models.Campaign{}
	response = models.Response{}

	if err := c.Bind(&updateCampaign); err != nil {
		response.Status = models.StatusError
		response.Message = err.Error()
		return c.JSON(getStatusCode(err), response)
	}

	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, updateCampaign)
	requestLogger.Info("Start to update a campaign.")
	err := cmpgn.CampaignUseCase.UpdateCampaign(c, c.Param("id"), &updateCampaign)

	if err != nil {
		response.Status = models.StatusError
		response.Message = err.Error()
		return c.JSON(getStatusCode(err), response)
	}

	response.Status = models.StatusSuccess
	response.Message = models.MessageUpdateSuccess
	requestLogger.Info("End of update a campaign.")

	return c.JSON(getStatusCode(err), response)
}

// GetCampaigns to get list of campaigns data
func (cmpgn *CampaignsHandler) GetCampaigns(c echo.Context) error {
	response = models.Response{}
	payload := map[string]interface{}{
		"name":      c.QueryParam("name"),
		"status":    c.QueryParam("status"),
		"startDate": c.QueryParam("startDate"),
		"endDate":   c.QueryParam("endDate"),
		"page":      c.QueryParam("page"),
		"limit":     c.QueryParam("limit"),
	}

	logger := models.RequestLogger{
		Payload: payload,
	}

	requestLogger := logger.GetRequestLogger(c, nil)
	requestLogger.Info("Start to get campaigns.")
	countCampaign, data, err := cmpgn.CampaignUseCase.GetCampaign(c, payload)

	if err != nil {
		response.Status = models.StatusError
		response.Message = err.Error()
		return c.JSON(getStatusCode(err), response)
	}

	if len(data) > 0 {
		response.Data = data
	}

	response.Status = models.StatusSuccess
	response.Message = models.MessageDataSuccess
	response.TotalCount = countCampaign

	requestLogger.Info("End of get campaigns.")

	return c.JSON(http.StatusOK, response)
}

// GetCampaignDetail a handler  to provide and endpoint to get campaign detail
func (cmpgn *CampaignsHandler) GetCampaignDetail(c echo.Context) error {
	response = models.Response{}
	logger := models.RequestLogger{}
	requestLogger := logger.GetRequestLogger(c, nil)
	requestLogger.Info("Start to get detail campaign.")

	responseData, err := cmpgn.CampaignUseCase.GetCampaignDetail(c, c.Param("id"))

	if err != nil {
		response.Status = models.StatusError
		response.Message = err.Error()
		return c.JSON(getStatusCode(err), response)
	}

	if (&models.Campaign{}) != responseData {
		response.Data = responseData
	}

	response.Status = models.StatusSuccess
	response.Message = models.MessagePointSuccess
	requestLogger.Info("End of get detail campaign.")

	return c.JSON(http.StatusOK, response)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if strings.Contains(err.Error(), "400") {
		return http.StatusBadRequest
	}

	switch err {
	case models.ErrInternalServerError:
		return http.StatusInternalServerError
	case models.ErrNotFound:
		return http.StatusNotFound
	case models.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusOK
	}
}
