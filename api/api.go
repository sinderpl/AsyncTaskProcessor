package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type server struct {
	listenAddr string
}

type errorResponse struct {
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

// Run starts the serve and listens on the specified port
func (s *server) Run() error {
	router := mux.NewRouter()

	router.Handle("/healthz", makeHTTPHandleFunc(s.handleHealthz)).
		Methods(http.MethodGet)

	router.
		HandleFunc("/tasks/enqueue", makeHTTPHandleFunc(s.handleTaskEnqueue)).
		Methods(http.MethodPost)

	router.
		HandleFunc("/task/{id}", makeHTTPHandleFunc(s.handleGetTaskInfo)).
		Methods(http.MethodGet)

	router.
		HandleFunc("/tasks/", makeHTTPHandleFunc(s.handleGetTasksInfo)).
		Methods(http.MethodPost)

	// TODO get tasks by status

	http.ListenAndServe(s.listenAddr, router)

	return nil
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) error {
	return writeJson(w, http.StatusOK, "service is healthy")
}

func (s *server) handleTaskEnqueue(w http.ResponseWriter, r *http.Request) error {
	panic("implement me")
}

func (s *server) handleGetTaskInfo(w http.ResponseWriter, r *http.Request) error {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		return fmt.Errorf("id required to find task")

	}

	// TODO hook in persistence
	return writeJson(w, http.StatusOK, idStr)
}

func (s *server) handleGetTasksInfo(w http.ResponseWriter, r *http.Request) error {
	idStr := r.URL.Query().Get("ids")
	fmt.Println(idStr)

	return writeJson(w, http.StatusOK, idStr)
}

func makeHTTPHandleFunc(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if err := writeJson(w, http.StatusBadRequest, errorResponse{Error: err.Error()}); err != nil {
				log.Print(err)
			}
		}
	}
}

func writeJson(w http.ResponseWriter, status int, v any) error {

	// Setting headers after w.WriteHeader leads to these being ignored
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}
