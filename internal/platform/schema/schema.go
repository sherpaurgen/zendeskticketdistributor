package schema

import (
	"encoding/json"
	"log"
	"os"
	"ticket/cmd/zendesk"
)

type TicketList struct {
Results []Ticket `json:"results"`
}
type Ticket struct {
	Id               int64         `json:"id,omitempty"`
	CreatedAt       string     `json:"created_at,omitempty"`
	UpdatedAt       string     `json:"updated_at,omitempty"`
	Subject         string        `json:"subject,omitempty"`
	Description     string        `json:"description,omitempty"`
	Priority        string        `json:"priority,omitempty"`
	Status          string        `json:"status"`
	AssigneeId      int64           `json:"assignee_id,omitempty"`
	OrganizationId  int64           `json:"organization_id,omitempty"`
	Tags            []string      `json:"tags,omitempty"`
	}


type Agent struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Shift           string     `json:"shift"`
	Id             int64  `json:"id"`
	OrganizationId int64  `json:"organization_id"`
	Bias           int     `json:"bias"`
	Active         bool   `json:"active"`
	Suspended      bool   `json:"suspended"`

}
type AgentList struct {
	Users []Agent `json:"users"`
	//ingoring previous and next page
}

type TemplateResult struct {
	Results []NewTicket `json:"results"`
	Count        int         `json:"count"`
}
/* used for new ticket scanning in periodicjob go file  */
type NewTicket struct {
	Id         		int64         `json:"id"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
	Subject         string        `json:"subject"`
	Description     string        `json:"description"`
	Priority        string        `json:"priority"`
	Status          string        `json:"status"`
	AssigneeId      int64         `json:"assignee_id"`
	OrganizationId  int64         `json:"organization_id"`
	Updated_Stamp   string      `json:"updated_stamp"`
}

/* template used for new ticket assignment start  */
type TemplateAssignment struct {
TempTicket TaTicket  `json:"ticket"`
}

type TaTicket struct {
Status       string    `json:"status"`
UpdatedStamp string `json:"updated_stamp"`
AssigneeId   int64     `json:"assignee_id"`
SafeUpdate   bool      `json:"safe_update"`
Comment      TaComment `json:"comment"`
}
type TaComment struct {
	HtmlBody string `json:"html_body"`
}
/* template used for new ticket assignment end  */

/* used to set api endpoint of zendesk  */
func GetSubdomain() string{
	agentListFile:="/.triage/agents.json"
	var basePath string
	basePath = zendesk.GetUserHomeDir()
	agentjsonFilePath := basePath + agentListFile
	agentdata, err := os.ReadFile(agentjsonFilePath)
	if err != nil {
		log.Fatalf("Error reading agent config file ~/.triage/agents.json, expected format:\n { \"agentlist\":[\"kdb@example.com\",\"mhz@example.com\",\"bsl@example.com\"]}\n\n", err)
	}
	agentnames:= agentjson{Agentlist: []string{},Domain: "" }
	err=json.Unmarshal(agentdata, &agentnames)
	if err != nil {
		log.Fatalf("Problem unmarshalling ~/.triage/agents.json file")
	}
	log.Println("Reading agentlist json:",agentnames.Agentlist)
	return agentnames.Domain
}
type agentjson struct {
	Agentlist []string `json:"agentlist"`
	Domain string `json:"subdomain"`
}
