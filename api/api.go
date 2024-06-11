package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type server struct {
	listenAddr string
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

type ServerOption func(server *server)

// CreateApiServer creates and returns the server with predefined options
func CreateApiServer(opts ...ServerOption) *server {
	srv := server{listenAddr: ""}

	for _, opt := range opts {
		opt(&srv)
	}

	return &srv
}

// WithListenAddr sets the address the server should listen on
func WithListenAddr(addr string) ServerOption {
	return func(srv *server) {
		srv.listenAddr = addr
	}
}

func (s *server) Run() error {
	router := mux.NewRouter()

	router.
		HandleFunc("/tasks/enqueue", makeHTTPHandleFunc(s.handleTaskEnqueue)).
		Methods(http.MethodPost)

	router.
		HandleFunc("/task/{id}", makeHTTPHandleFunc(s.handleGetTaskInfo)).
		Methods(http.MethodGet)

	http.ListenAndServe(s.listenAddr, router)

	return nil
}

func (s *server) handleTaskEnqueue(w http.ResponseWriter, r *http.Request) error {
	panic("implement me")
}

func (s *server) handleGetTaskInfo(w http.ResponseWriter, r *http.Request) error {
	panic("implement me")
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if err := WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()}); err != nil {
				log.Print(err)
			}
		}
	}
}

func WriteJson(w http.ResponseWriter, status int, v any) error {

	// Setting headers after w.WriteHeader leads to these being ignored
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}
