package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	pingJSON := fmt.Sprintf("{\"status\":\"Success\",\"message\":\"PONG!!\"}\n")

	// Assertions
	if assert.NoError(t, ping(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, pingJSON, rec.Body.String())
	}
}
