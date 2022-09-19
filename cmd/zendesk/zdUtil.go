package zendesk

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io"
	"log"
	"net/http"
	"sync"
	"ticket/internal/platform/schema"
	"time"
)

// get open tickets of user
func GetUserOpenTicket(method string, secret string, querystr string, email string, db *sqlx.DB, wg *sync.WaitGroup) error {
	tgturl := schema.GetSubdomain() + "search.json?query=" + querystr
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, tgturl, nil)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Set("user-agent", "Firefox")
	req.Header.Set("Authorization", "basic "+secret)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	openticketlist := schema.TicketList{Results: []schema.Ticket{}}
	body, err := io.ReadAll(response.Body)
	if err := json.Unmarshal(body, &openticketlist); err != nil {
		log.Printf("Error unmarshalling json response: %v\n", err)
	}
	for _, v := range openticketlist.Results {
		//fmt.Printf("%v %v %v %v %v %v %v \n",v.Id,v.CreatedAt,v.UpdatedAt,v.Subject,v.Priority,v.Status,email)
		db.MustExec("INSERT INTO zd_tkt (tid,email,crtat,updat,sub,description,priority,status) VALUES ($1, $2, $3,$4,$5,$6,$7,$8)",
			v.Id, email, v.CreatedAt, v.UpdatedAt, v.Subject, v.Description, v.Priority, v.Status)
	}
	//log.Printf("%s Total: %d",email,len(openticketlist.Results))
	defer response.Body.Close()
	defer wg.Done()
	return nil
}


func GetUsers(method string, secret string, querystr string, db *sqlx.DB, wg *sync.WaitGroup) error {
	tgturl := schema.GetSubdomain()+"users/search.json?query=" + querystr
	log.Println("Fetching info for agent:", querystr)
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, tgturl, nil)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	req.Header.Set("user-agent", "Firefox")
	req.Header.Set("Authorization", "basic "+secret)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Got error %s", err.Error())
	}
	//manually setting default bias to zero and shift to evening
	userlistjson := schema.AgentList{Users: []schema.Agent{{Bias: 0, Shift: "OD"}}}
	body, err := io.ReadAll(response.Body)
	if err := json.Unmarshal(body, &userlistjson); err != nil {
		log.Printf("Error unmarshalling json response: %v\n", err)
	}
	//https://{subdomain}.zendesk.com/api/v2/users/search.json?query=sherpa@example.com -- it must return single obj in list
	if len(userlistjson.Users) != 1 {
		log.Fatalln("Issue in GetUsers function while fetching user info: please check agent email json file")
	}

	db.MustExec("INSERT INTO agent (aid, name,email,orgid , active ,suspnd ,bias ,shift) VALUES ($1, $2, $3,$4,$5,$6,$7,$8)",
		userlistjson.Users[0].Id, userlistjson.Users[0].Name, userlistjson.Users[0].Email, userlistjson.Users[0].OrganizationId, userlistjson.Users[0].Active, userlistjson.Users[0].Suspended, userlistjson.Users[0].Bias, userlistjson.Users[0].Shift)

	//log.Printf("%s Total: %d",email,len(openticketlist.Results))
	defer response.Body.Close()
	defer wg.Done()
	return nil
}
