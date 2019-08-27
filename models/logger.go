package models

import (
	"encoding/json"
	"errors"

	"github.com/fatih/structs"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// RequestLogger a struct to store a request logger
type RequestLogger struct {
	RequestID string      `json:"requestID,required"`
	Payload   interface{} `json:"payload,omitempty"`
}

var errReqLogger = errors.New("Error during creating a request logger")

// GetRequestLogger is to get a log parameter
func (rl *RequestLogger) GetRequestLogger(c echo.Context, payload interface{}) *logrus.Entry {
	rl.RequestID = c.Response().Header().Get(echo.HeaderXRequestID)
	pl, err := json.Marshal(payload)

	if err != nil {
		logrus.WithFields(logrus.Fields{"requestID": rl.RequestID}).Debug(err)

		return logrus.WithFields(logrus.Fields{})
	}

	if payload != nil {
		rl.Payload = string(pl)
	}

	return logrus.WithFields(logrus.Fields{"params": structs.Map(rl)})
}

// DataLog is to get a log parameter
func (rl *RequestLogger) DataLog(c echo.Context, payload interface{}) *logrus.Entry {
	rl.RequestID = c.Response().Header().Get(echo.HeaderXRequestID)
	pl, err := json.Marshal(payload)

	if err != nil {
		logrus.WithFields(logrus.Fields{"requestID": rl.RequestID}).Debug(err)

		return logrus.WithFields(logrus.Fields{})
	}

	if payload != nil {
		rl.Payload = string(pl)
	}

	return logrus.WithFields(logrus.Fields{"params": structs.Map(rl)})
}
