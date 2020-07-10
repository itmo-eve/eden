package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lf-edge/eden/eserver/pkg/manager"
	"log"
	"net/http"
)

type EServer struct {
	Port    string
	Address string
	Manager *manager.EServerManager
}

//Start http server
//  /admin/list endpoint returns list of files
//  /admin/add-from-url endpoint fires download
//  /admin/status/{filename} returns fileinfo
//  /eserver/{filename} returns file
func (s *EServer) Start() {

	s.Manager.Init()

	api := &apiHandler{
		manager: s.Manager,
	}

	admin := &adminHandler{
		manager: s.Manager,
	}

	router := mux.NewRouter()

	ad := router.PathPrefix("/admin").Subrouter()

	ad.HandleFunc("/list", admin.list).Methods("GET")
	ad.HandleFunc("/add-from-url", admin.addFromUrl).Methods("POST")
	ad.HandleFunc("/status/{filename}", admin.getFileStatus).Methods("GET")

	router.HandleFunc("/eserver/{filename}", api.getFile).Methods("GET")

	server := &http.Server{
		Handler: router,
		Addr:    fmt.Sprintf("%s:%s", s.Address, s.Port),
	}

	log.Println("Starting eserver:")
	log.Printf("\tIP:Port: %s:%s\n", s.Address, s.Port)
	log.Printf("\tDirectory: %s\n", s.Manager.Dir)
	log.Fatal(server.ListenAndServe())
}
