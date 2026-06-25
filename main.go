package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/micaroni/risk-weaver/internal/routes"
	workload "github.com/micaroni/risk-weaver/internal/workloadhandler"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("server failure: ", err.Error())
		os.Exit(1)
	}
}

func run() error {
	workloadServ := workload.InitWorkloadService(WLRequest)

	mux := routes.InitRoutes(workloadServ)

	serv := http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: mux,
	}
	return serv.ListenAndServe()
}
