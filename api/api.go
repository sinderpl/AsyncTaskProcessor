package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sinderpl/AsyncTaskProcessor/task"
)

const testUserId = "testUserIdTodo"

type server struct {
	listenAddr string
	taskChan   *chan []*task.Task
}

type errorResponse struct {
	Priority task.ExecutionPriority `json:"priority,omitempty"`
	TaskType task.TypeOf            `json:"taskType,omitempty"`
	Error    string                 `json:"error" json:"error"`
}

type option func(server *server)

// CreateApiServer creates and returns the server with predefined options
func CreateApiServer(opts ...option) *server {
	srv := server{listenAddr: ""}

	for _, opt := range opts {
		opt(&srv)
	}

	return &srv
}

// WithListenAddr *required* sets the address the server should listen on
func WithListenAddr(addr string) option {
	return func(srv *server) {
		srv.listenAddr = addr
	}
}

// WithQueue *required* the queue will listen to new tasks on this chan
func WithQueue(taskChan *chan []*task.Task) option {
	return func(q *server) {
		q.taskChan = taskChan
	}
}

// Run starts the serve and listens on the specified port
func (s *server) Run() error {
	router := mux.NewRouter()

	// TODO add middleware to validate user / api key
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

	// TODO get Tasks by status

	fmt.Println("server ready  and listening for requests")
	http.ListenAndServe(s.listenAddr, router)

	return nil
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) error {
	return writeJson(w, http.StatusOK, "service is healthy")
}

func (s *server) handleTaskEnqueue(w http.ResponseWriter, r *http.Request) error {

	req := new(EnqueueTaskPayload)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode request body", err)
		return errors.New("failed to decode request body")
	}

	resp := EnqueueTaskResponse{
		Tasks: make([]TaskResponse, 0, len(req.Tasks)),
	}

	newTasks := make([]*task.Task, 0, len(req.Tasks))

	for _, t := range req.Tasks {
		var tResp TaskResponse

		newTask, err := task.CreateTask(
			task.WithType(t.TaskType),
			task.WithCreatedBy(testUserId), // TODO add user session validation
			task.WithPriority(t.Priority),
			task.WithPayload(t.Payload))

		if newTask == nil || err != nil {
			resp := errorResponse{
				Priority: t.Priority,
				TaskType: t.TaskType,
				Error:    fmt.Sprintf("failed to create task: %v", err),
			}
			return writeJson(w, http.StatusBadRequest, resp)
		}

		tResp = TaskResponse{
			Id:       newTask.Id,
			TaskType: newTask.Type,
			Priority: newTask.Priority,
			Status:   newTask.Status,
		}

		newTasks = append(newTasks, newTask)
		resp.Tasks = append(resp.Tasks, tResp)
	}

	// Write tasks to queue so it can distribute and begin processing
	*s.taskChan <- newTasks

	resp.Status = "Successfully enqueued valid tasks"

	return writeJson(w, http.StatusOK, resp)
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
			// TODO extend these to provide better http codes from function returns
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

	fmt.Println(v)

	return json.NewEncoder(w).Encode(v)
}
