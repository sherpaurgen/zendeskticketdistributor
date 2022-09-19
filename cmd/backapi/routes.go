package main

import (
	"github.com/go-chi/chi/v5"
    "github.com/go-chi/cors"
)
// receiver fxn below, exporting all routes from here
func (app *Application) routes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS","PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	router.Get("/api/v2",app.statusHandler)
	router.Get("/api/v2/listot",app.listOT)
	router.Get("/api/v2/listpt",app.listPT)
	router.Get("/api/v2/gettc",app.getTcount)
	router.Get("/api/v2/agentlist",app.AgentList)
	router.Get("/api/v2/resyncdb",app.Resyncdb)
	router.Get("/api/v2/refreshticket",app.refreshticket)
	router.Get("/api/v2/stats",app.AgentStat)
	router.Patch("/api/v2/setbias/{id}",app.setBias)
	router.Patch("/api/v2/setShift/{id}",app.setShift)
return router
}
