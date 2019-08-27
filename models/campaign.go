package models

import (
	"time"
)

var (
	// CampaignInActive to store if campaign not active
	CampaignInActive int64

	// CampaignActive to store  if campaign active
	CampaignActive int64 = 1
)

// Campaign is represent a campaigns model
type Campaign struct {
	ID          int64      `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	EndDate     string     `json:"endDate,omitempty"`
	StartDate   string     `json:"startDate,omitempty"`
	Status      *int8      `json:"status,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	Rewards     *[]Reward  `json:"rewards,omitempty"`
}

// GetCampaignValue to store payload get campaign value
type GetCampaignValue struct {
	CIF               string  `json:"cif,omitempty"`
	Channel           string  `json:"channel,omitempty"`
	Product           string  `json:"product,omitempty"`
	TransactionType   string  `json:"transactionType,omitempty"`
	Unit              string  `json:"unit,omitempty"`
	TransactionAmount float64 `json:"transactionAmount,omitempty"`
	Source            string  `json:"source,omitempty"` // device name that user used
	RefCore           string  `json:"refCore,omitempty"`
}
