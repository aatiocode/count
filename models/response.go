package models

import (
	"strings"
)

// Response struct is represent a data for output payload
type Response struct {
	Code        string      `json:"code,omitempty"`
	Status      string      `json:"status,omitempty"`
	Message     string      `json:"message,omitempty"`
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	TotalCount  string      `json:"totalCount,omitempty"`
}

// ResponseErrors struct is represent a data for output payload
type ResponseErrors struct {
	Title   string
	Details []string
}

var (
	responseSuccess = "Success"
	responseError   = "Error"
)

var responseCode = map[string]string{
	responseSuccess: "00",
	responseError:   "99",
}

// SetTitle title of Response errors
func (re *ResponseErrors) SetTitle(title string) {
	re.Title = title
}

// AddError adding error on Response errors
func (re *ResponseErrors) AddError(errString string) {
	re.Details = append(re.Details, errString)
}

// SetResponse to bla bla
func (resp *Response) SetResponse(respData interface{}, respErrors *ResponseErrors) {
	if respErrors != nil && respErrors.Title != "" {
		resp.Status = responseError
		resp.Code = responseCode[responseError]
		resp.Message = respErrors.Title
		if len(respErrors.Details) != 0 {
			resp.Description = strings.Join(respErrors.Details, ", ")
		}
	} else {
		resp.Status = responseSuccess
		resp.Code = responseCode[responseSuccess]
		resp.Message = MessageDataSuccess
		resp.Data = respData
	}
}
