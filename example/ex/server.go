package ex

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/benschw/opin-go/ophttp"
	"github.com/benschw/opin-go/rest"
	"github.com/gorilla/mux"
)

type Handler struct {
	Stats   map[string][]string
	Request *http.Request
}

func (h *Handler) addStat(key string, body string) {
	if h.Stats[key] == nil {
		h.Stats[key] = make([]string, 0)
	}
	h.Stats[key] = append(h.Stats[key], body)
}

func (h *Handler) Foo(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)
	log.Printf("HTTP: Foo: '%s'", s)

	h.addStat("Foo", s)
	h.Request = req
	rest.SetOKResponse(res, nil)
}
func (h *Handler) Bar(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)
	log.Printf("HTTP: Bar: '%s'", s)

	h.addStat("Bar", s)
	rest.SetOKResponse(res, nil)
}
func (h *Handler) Bad(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)
	log.Printf("HTTP: Bad: '%s'", s)

	h.addStat("Bad", s)
	rest.SetInternalServerErrorResponse(res, fmt.Errorf("Setting Error"))
}
func (h *Handler) Slow(res http.ResponseWriter, req *http.Request) {
	bs, _ := ioutil.ReadAll(req.Body)
	s := string(bs)
	log.Printf("HTTP: Slow: '%s'", s)

	time.Sleep(5000 * time.Millisecond)

	h.addStat("Slow", s)
	rest.SetOKResponse(res, nil)
}

// Run The Server
func NewServer(bind string) *Server {
	server := ophttp.NewServer(bind)
	stats := make(map[string][]string)
	return &Server{
		S: server,
		H: &Handler{Stats: stats},
	}
}

type Server struct {
	S *ophttp.Server
	H *Handler
}

func (s *Server) Start() error {
	log.Println("HTTP: Starting Demo Http Server")

	r := mux.NewRouter()

	r.HandleFunc("/foo", s.H.Foo).Methods("POST")
	r.HandleFunc("/bar", s.H.Bar).Methods("POST")
	r.HandleFunc("/bad", s.H.Bad).Methods("POST")
	r.HandleFunc("/slow", s.H.Slow).Methods("POST")

	sMux := http.NewServeMux()
	sMux.Handle("/", r)

	return s.S.Start(sMux)
}
func (s *Server) Stop() {
	s.S.Stop()
}
