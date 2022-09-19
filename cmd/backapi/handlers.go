package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"ticket/cmd/zendesk"
	"ticket/internal/platform/schema"
	"time"
)

type ticketCountStatus struct {
	Name            string `db:"name"`
	Email           string `db:"email"`
	Status          string `db:"status"`
	Count           int 	`db:"count"`
}
func (app *Application) setBias(w http.ResponseWriter,r *http.Request){
	type BiasVal struct {
		Bias int `json:"bias"`
	}
	bv:=BiasVal{}
	agentid:=chi.URLParam(r,"id")
	body, err := io.ReadAll(r.Body)
	//app.log.Printf("msg body %s",body)
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
    if err!=nil {
		body:=`{"mesg":"Cannot read req body bias value"}`
		js,_:=json.Marshal(body)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(js)
	}
	if err := json.Unmarshal(body, &bv); err != nil {
		app.log.Printf("Error unmarshalling bias value json response: %v\n", err)
	}
	app.log.Println("agentid: ",agentid)
	app.log.Println("thePatchBody: ",bv.Bias)
}

func (app *Application) setShift(w http.ResponseWriter,r *http.Request){
	type ShiftVal struct {
		Shift string `json:"shift"`
	}
	sv:=ShiftVal{}
	agentid:=chi.URLParam(r,"id")
	body, err := io.ReadAll(r.Body)
	//app.log.Printf("msg body %s",body)
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err!=nil {
		body:=`{"mesg":"Cannot read req body shift value"}`
		js,_:=json.Marshal(body)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(js)
	}
	if err := json.Unmarshal(body, &sv); err != nil {
		app.log.Printf("Error unmarshalling shift value json response: %v\n", err)
	}
	newshift:=""
	if sv.Shift=="OD" {
		newshift="Off"
	} else if sv.Shift=="Off" {
		newshift="OD"
	} else {
		newshift="Off"
	}
	app.db.MustExec("UPDATE agent SET shift = $1 WHERE aid = $2",newshift,agentid)
	//app.log.Println("agentid: ",agentid)
	//app.log.Println("theSetshiftPatchBody: ", sv.Shift)
	message:=`{"mesg":"shift changed"}`
	js,_:=json.Marshal(message)
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}

func (app *Application) statusHandler(w http.ResponseWriter,r *http.Request) {
	body:=`{"mesg":"hello World"}`
	js,err:=json.Marshal(body)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
func (app *Application) refreshticket(w http.ResponseWriter,r *http.Request) {
	var wgrt sync.WaitGroup
	app.db.MustExec("TRUNCATE table zd_tkt")
	for _, val := range app.AgentArray{
		wgrt.Add(1)
		query := "assignee:" + val + " status:open status:pending"
		url.QueryEscape(val)
		go zendesk.GetUserOpenTicket("GET", zendesk.GetZdConfig(), url.QueryEscape(query), val, app.db, &wgrt)
	}
	body:=`{"mesg":"updating open_pending ticket list"}`
	js,err:=json.Marshal(body)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
	app.log.Println("refreshticketCalled..")
}


func (app *Application) listOT(w http.ResponseWriter,r *http.Request) {
    ot:=[]ticketCountStatus{}  //openticket and openticketstruct LIST
	err:=app.db.Select(&ot, "SELECT email,status,COUNT(email) from zd_tkt WHERE status='open' group by email,status")
	//SELECT COUNT(email),email,status from zd_tkt WHERE status='open' group by email,status;
	js,err:=json.Marshal(ot)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
		w.Header().Set("Content-Type","application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(js)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
//get pending tickets of users
func (app *Application) listPT(w http.ResponseWriter,r *http.Request) {
	pt:=[]ticketCountStatus{}  //openticket and openticketstruct LIST
	err:=app.db.Select(&pt, "SELECT email,status,COUNT(email) from zd_tkt WHERE status='pending' group by email,status")
	js,err:=json.Marshal(pt)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
		w.Header().Set("Content-Type","application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		js=[]byte(`{"msg":"Error encoding json"}`)
		w.Write(js)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
//get TOTAL ticket COUNT of users
func (app *Application) getTcount(w http.ResponseWriter,r *http.Request) {
	tc:=[]ticketCountStatus{}  //openticket and openticketstruct LIST
	//err:=app.db.Select(&tc, "SELECT email,COUNT(email) FROM zd_tkt GROUP BY email;")
	err:=app.db.Select(&tc, "SELECT agent.name,zd_tkt.email,COUNT(*) \nFROM \n    zd_tkt\nINNER JOIN agent \n  ON zd_tkt.email=agent.email\n  GROUP BY zd_tkt.email,agent.name ORDER BY count ASC")
	js,err:=json.Marshal(tc)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusInternalServerError)
		js=[]byte(`{"msg":"Error encoding json"}`)
		w.Write(js)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
//get agents list - shift info-night,morning etc and ,name,email etc
type Agentdata struct {
	Name string
	Email string
	Shift string
	Aid int64
	Orgid int64
	Bias int
	Suspnd bool
	Active bool
}

func (app *Application) AgentList(w http.ResponseWriter,r *http.Request) {
	agents:=[]Agentdata{}  //openticket and openticketstruct LIST
	err:=app.db.Select(&agents, "select aid,name,email,orgid,active,suspnd,bias,shift from agent")
	js,err:=json.Marshal(agents)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
		w.Header().Set("Content-Type","application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		js=[]byte(`{"msg":"Error encoding json"}`)
		w.Write(js)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
func (app *Application) Resyncdb(w http.ResponseWriter,r *http.Request) {
	//this will truncate the zd_tkt table which have all open/pending tickets data
	_,err:=app.db.Exec("TRUNCATE table zd_tkt")
	if err!=nil{
		w.Header().Set("Content-Type","application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		js:=[]byte(`{"msg":"Error encoding json"}`)
		w.Write(js)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	js:=[]byte(`{"msg":"DB is resynced"}`)
	w.Write(js)

	for _, val := range app.AgentArray {
		email:=val
	    query := "assignee:" + val + " status:open status:pending"
		secret:= zendesk.GetZdConfig()
		tgturl := schema.GetSubdomain()+"search.json?query=" + url.QueryEscape(query)
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest("GET", tgturl, nil)
		if err != nil {
			app.log.Println("Error connecting to zd from resync db handler:",err)
		}
		req.Header.Set("user-agent", "Firefox")
		req.Header.Set("Authorization", "basic "+secret)
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		log.Printf("resyncdb response status %v\n",response.StatusCode)

		if err != nil {
			app.log.Println("Error connecting to zd from resync db handler:",err)
		}
		openticketlist := schema.TicketList{Results: []schema.Ticket{{Priority: "normal"}}}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err := json.Unmarshal(body, &openticketlist); err != nil {
			app.log.Printf("Resyncdb: Error unmarshalling json response: %v\n", err)
		}
		for _, v := range openticketlist.Results {

			//fmt.Printf("%v %v %v %v %v %v %v \n",v.Id,v.CreatedAt,v.UpdatedAt,v.Subject,v.Priority,v.Status,email)
			app.db.MustExec("INSERT INTO zd_tkt (tid,email,crtat,updat,sub,description,priority,status) VALUES ($1, $2, $3,$4,$5,$6,$7,$8)",
				v.Id, email, v.CreatedAt, v.UpdatedAt, v.Subject, v.Description, v.Priority, v.Status)
		}
		//log.Printf("%s Total: %d",email,len(openticketlist.Results))
	}
}

func (app *Application) AgentStat(w http.ResponseWriter,r *http.Request) {
	stats:=[]TicketAssignedStats{}
	err:=app.db.Select(&stats, "SELECT COUNT(email),email FROM tcounter WHERE updat > (CURRENT_DATE - INTERVAL '7 days') GROUP BY email ORDER BY COUNT DESC;")
	js,err:=json.Marshal(stats)
	if err!=nil{
		app.log.Println("Error encoding json:",err)
		w.Header().Set("Content-Type","application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusInternalServerError)
		js=[]byte(`{"msg":"Error encoding json"}`)
		w.Write(js)
	}
	w.Header().Set("Content-Type","application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}

type TicketAssignedStats struct {
	Count int      `db:"count"`
	Email string `db:"email"`
}