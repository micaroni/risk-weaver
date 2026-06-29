package routes

import (
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	workload "github.com/micaroni/risk-weaver/internal/workloadhandler"
)

var (
	routeSyncOne sync.Once
	router       *httprouter.Router
)

const (
	RouteWorkload = "/workloads"
)

func InitRoutes(ws workload.WorkloadService) http.Handler {
	routeSyncOne.Do(func() {
		router = initRoutes(ws)
	})

	return router
}

func initRoutes(ws workload.WorkloadService) *httprouter.Router {
	mux := httprouter.New()
	mux.HandleOPTIONS = true

	initWorkloadRoutes(mux, ws)
	return mux
}

func initWorkloadRoutes(mux *httprouter.Router, ws workload.WorkloadService) {
	mux.POST(RouteWorkload, ws.AddNewWorkload())
	mux.GET(RouteWorkload, ws.RetrieveWorkload())
}
