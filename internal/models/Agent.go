package models

import "time"

type Agent struct {
		Id                   int64         `json:"id"`
		Url                  string        `json:"url"`
		Name                 string        `json:"name"`
		Email                string        `json:"email"`
		OrganizationId       int64         `json:"organization_id"`
		Active               bool          `json:"active"`
		Suspended            bool          `json:"suspended"`
		Bias                 int           `json:"Bias"`
		Shift               bool		   `json:"OnDuty"`
		Counter              int           `json:"counter"`
}

type	Tkt struct {
	Id              int           `json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Subject         string        `json:"subject"`
	Description     string        `json:"description"`
	Priority        string        `json:"priority"`
	Status          string        `json:"status"`
	OrganizationId  int64         `json:"organization_id"`
	Tags            []string      `json:"tags"`
}





