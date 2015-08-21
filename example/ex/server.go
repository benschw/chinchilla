package ex

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/benschw/opin-go/ophttp"
	"github.com/benschw/opin-go/rest"
	"github.com/gorilla/mux"
)

func FooHandler(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)

	log.Printf("HTTP: Foo: '%s'", s)
	rest.SetOKResponse(res, nil)
}
func BarHandler(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)

	log.Printf("HTTP: Bar: '%s'", s)
	rest.SetOKResponse(res, nil)
}
func BadHandler(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)

	log.Printf("HTTP: Bad: '%s'", s)
	rest.SetInternalServerErrorResponse(res, fmt.Errorf("Setting Error"))
}

// Run The Server
func NewServer(bind string) *Server {
	server := ophttp.NewServer(bind)
	return &Server{S: server}
}

type Server struct {
	S *ophttp.Server
}

func (s *Server) Start() error {
	log.Println("HTTP: Starting Demo Http Server")
	r := mux.NewRouter()

	r.HandleFunc("/foo", FooHandler).Methods("POST")
	r.HandleFunc("/bar", BarHandler).Methods("POST")
	r.HandleFunc("/bad", BadHandler).Methods("POST")

	http.Handle("/", r)

	return s.S.Start()
}
func (s *Server) Stop() {
	s.S.Stop()
}
