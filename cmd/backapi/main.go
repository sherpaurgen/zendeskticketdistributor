package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"ticket/cmd/backapi/errands"
	"ticket/cmd/zendesk"
	"ticket/internal/platform/database"
	"time"
)

type Application struct {
	log *log.Logger
	db  *sqlx.DB
	AgentArray []string
	wg *sync.WaitGroup
	ch chan error
}

func main() {

	logger := log.New(os.Stdout, "HTTPServer ", log.Ldate|log.Ltime)
	logger.Println("Starting main...")
	var wg sync.WaitGroup
	db := database.InitDBConn()
	err := db.Ping()

	if err != nil {
		log.Fatalln("Main db connection err:", err)
	}
	agentlist:=zendesk.GetAgentList()
	app := Application{
		log: logger,
		db:  db,
		AgentArray: agentlist.Agentlist,
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", "127.0.0.1", "8080"),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

		//truncating tables to fetch fresh data from zendeskapi
		db.MustExec("TRUNCATE table zd_tkt")
		db.MustExec("TRUNCATE table agent")
		for _, val := range agentlist.Agentlist {
			wg.Add(1)
			query := "assignee:" + val + " status:open status:pending"
			url.QueryEscape(val)
			go zendesk.GetUserOpenTicket("GET", zendesk.GetZdConfig(), url.QueryEscape(query), val, db, &wg)
		}
		for _, val := range agentlist.Agentlist {
			wg.Add(1)
			go zendesk.GetUsers("GET", zendesk.GetZdConfig(), val, db, &wg)
		}

		query := "status:new tags:paying_customer"
		wg.Add(1)

		go errands.TicketScanner("GET", zendesk.GetZdConfig(), url.QueryEscape(query), db, &wg)
		wg.Add(1)
		go errands.RefreshTicketList(&wg)

	log.Println("Starting http server")
	err = srv.ListenAndServe()
	if err != nil {
		log.Println("Error starting http server")
	}
    //go errands.JobGetNewTicket()

	wg.Wait()
}
//go run  ./cmd/backapi

