package errands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"ticket/internal/platform/schema"
	"time"
)
type MsgResp struct {
	mesg string
}
type NumAgent struct {
	Count int `db:"count"`
}
type AgentTicketCount struct {
	Email string  `db:"email"`
	Aid int64      `db:"aid"`
	Count int   `db:"count"`
}

func RefreshTicketList(wg *sync.WaitGroup){
	ticker:=time.NewTicker(7*time.Minute)
	for _ = range ticker.C{
		//log.Println("Calling refreshticket from errands" to get open pending tickets )
		resp, err := http.Get("http://127.0.0.1:8080/api/v2/refreshticket")
		resp.Body.Close()
		if err != nil {
			log.Fatalln("Error accessing /api/v2/refreshticket")
			wg.Done()
		}

		//body, err := io.ReadAll(resp.Body)
		//log.Println(body)
	}
}
/* response example for getting list of new ticket
{"results":[],"facets":null,"next_page":null,"previous_page":null,"count":0}
*/
func TicketScanner(method string, secret string, querystr string, db *sqlx.DB, wg *sync.WaitGroup) error {
	ticker:=time.NewTicker(10*time.Minute)
	for _ = range ticker.C {
		//fetch new ticket here and save temporarily in struct
		tgturl := schema.GetSubdomain() + "search.json?query=" + querystr
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest(method, tgturl, nil)
		if err != nil {
			return fmt.Errorf("Got error in periodicjob %s", err.Error())
		}
		req.Header.Set("user-agent", "Firefox")
		req.Header.Set("Authorization", "basic "+secret)
		req.Header.Set("Content-Type", "application/json")
		log.Println("Sending http request...")
		response, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		//var openticketlist schema.TicketS
		body, err := io.ReadAll(response.Body)
		log.Println("Request resp status:",response.Status)
		err=response.Body.Close()
		if err!=nil{
			log.Fatalln("Problem closing resp",err)
		}
		newticketlist := schema.TemplateResult{Results: []schema.NewTicket{} }
		if err := json.Unmarshal(body, &newticketlist); err != nil {
			log.Printf("Error unmarshalling json response: %v\n", err)
		}
		NumOfNT:=len(newticketlist.Results)
		/* if no ticket found jump to ticker*/
		if NumOfNT==0{
			log.Println("len of newticketlist 0 NumofNT<1: going to Ticker ",len(newticketlist.Results))
			continue
		}
        isduplicate:=false
		for i,_ := range newticketlist.Results {
			if newticketlist.Results[i].Status=="open"{
				log.Printf("Stale result %v %v\n",newticketlist.Results[i].Id,newticketlist.Results[i].Status)
				isduplicate=true
				break
			}
		}
		if isduplicate{
			continue
		}

		var totalodagent int
		/*total agent on OD*/
		row := db.QueryRow("SELECT distinct count(email)::int FROM agent WHERE shift='OD'")
		err = row.Scan(&totalodagent)
		if err !=nil {
			log.Fatalln("Error getting total number of agents ONSHIFT periodicjob:",err)
		}

		atcount:=[]AgentTicketCount{}
		log.Printf("NumNewticket- %v , totalodagent: %v",NumOfNT,totalodagent)
		/* assignment logic here if number of ticket is bigger than available agents then slice
		the tickets and divide it to same number of agent in the agent list eg. if numticket=10 and agent=4 ,
		each agent will get 2 tickets(remaining 2 will be handled in next iteration same way)
		*/
		if NumOfNT>totalodagent {
			/* Get OnDuty agent with low number of open/pending ticket and insert into tcount struct */
			err = db.Select(&atcount,"SELECT agent.aid,zd_tkt.email,COUNT(*) \nFROM \n zd_tkt\nINNER JOIN agent \n  ON zd_tkt.email=agent.email\nWHERE agent.shift='OD'  GROUP BY zd_tkt.email,agent.aid ORDER BY count ASC LIMIT $1",totalodagent)
		  	if err !=nil {
				log.Fatalln("Error finding agent with few tasks at hand NumOfNT>totalodagent periodicjob:",err)
		  	}

			ticketslice:=newticketlist.Results[0:totalodagent]
			for i,_ := range ticketslice {
				log.Printf("_HigherT_ agent id %v gets ticketId: %v \n",atcount[i].Email,ticketslice[i].Id)
				err:=AssignTicket(secret,db,atcount[i].Email,ticketslice[i].Id,atcount[i].Aid)
				if err!=nil {
					log.Println("Error Assigning ticket: Ticketscanner")
				}
			}
			newticketlist=schema.TemplateResult{Results: nil}
			atcount=nil
		} else if NumOfNT<=totalodagent {
			log.Println("Inside second ifcase")
			err = db.Select(&atcount,"SELECT agent.aid,zd_tkt.email,COUNT(*) \nFROM \n zd_tkt\nINNER JOIN agent \n  ON zd_tkt.email=agent.email\nWHERE agent.shift='OD'  GROUP BY zd_tkt.email,agent.aid ORDER BY count ASC LIMIT $1",NumOfNT)
			if err !=nil {
				log.Fatalln("Error finding agent with few tasks at hand nt<=totalodagent periodicjob:",err)
			}

			for i,_:=range atcount {
				log.Printf("_LowerT_ agent id %v gets ticketId: %v",atcount[i].Email,newticketlist.Results[i].Id)
				err:=AssignTicket(secret , db ,atcount[i].Email,newticketlist.Results[i].Id ,atcount[i].Aid)
				if err!=nil {
					log.Println("Error Assigning ticket: Ticketscanner")
				}
			}

			newticketlist=schema.TemplateResult{Results: nil}
			atcount=nil

		} else {
			log.Println("----continuing to ticker..")
			continue
		}
	}
	defer wg.Done()
	return nil
}

func AssignTicket(secret string, db *sqlx.DB,email string,ticketid int64,agentid int64) error {
	//method should be put
	method:="PUT"
	tgturl := schema.GetSubdomain() + "tickets/"+ strconv.FormatInt(ticketid, 10)
	//log.Printf("%vhello",tgturl)
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	t:=schema.TemplateAssignment{ TempTicket: schema.TaTicket {
		Status: "open",
		AssigneeId: agentid,
		SafeUpdate: true,
		UpdatedStamp: GetCurTime(),
		Comment: schema.TaComment{HtmlBody: GetMessage() },
	    },
	}
	///log.Println("template data",t)
	json,err:=json.Marshal(t)
	//log.Println("template json:",string(json))
	if err !=nil { log.Println("Error encoding json Assignticket : ",err) }
	/* Sending put request to assign ticket */
	req, err := http.NewRequest(method, tgturl, bytes.NewBuffer(json))
	if err != nil {
		 log.Printf("Got error in periodicjob Assignticket %v \n", err.Error())
	}
	req.Header.Set("user-agent", "Firefox")
	req.Header.Set("Authorization", "basic "+secret)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Got error %v", err.Error())
	}
	log.Printf("Assigned ticket %v to %v\n",ticketid,email)
	//var openticketlist schema.TicketS
	 _, err = io.ReadAll(resp.Body)
	//log.Printf("Ticket Assign Response %v : %v %v\n :-",ticketid,resp.StatusCode,string(body))
	/* inserting ticket assignment to tcounter table postgres*/
	assignmentdata:=`INSERT INTO tcounter ( email,tid,updat ) VALUES ($1,$2,$3)`
	db.MustExec(assignmentdata,email,ticketid,time.Now())
	return nil
}

func GetCurTime() string {
	loc, _ := time.LoadLocation("UTC") // use other time zones such as MST, IST
	//get time in that zone
	t := time.Now().In(loc).Add(time.Minute * 1)
	Updated_Stamp := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d"+"Z",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
	return Updated_Stamp
}

func GetMessage() string {
	msg:="<p>Hello {{ticket.requester.name}}</p>Thanks for contacting us! I am reviewing your ticket , and will get back to you as soon as possible. " +
		"To add additional comments, reply to this email, or follow the ticket link.<p>"
    return msg
}