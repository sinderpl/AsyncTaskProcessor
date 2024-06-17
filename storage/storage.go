package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/sinderpl/AsyncTaskProcessor/task"
	"log"
	"log/slog"
)

// TODO add batch create task
type Storage interface {
	CreateTask(*task.Task) error
	UpdateTask(*task.Task) error
	GetTaskById(string) (*task.Task, error)
}

type PostgresStore struct {
	db   *sql.DB
	name string
}

func (p *PostgresStore) Init() error {
	return p.createTaskTable()
}

func (p *PostgresStore) createTaskTable() error {
	query := `create table if not exists tasks (
		id varchar(100) primary key,
		priority int,
    	taskType varchar(30),
    	status varchar(60),
    	backOffDuration bigint,
    	payload jsonb,      
        createdAt timestamp,
        createdBy varchar(30),
    	startedAt  timestamp,
    	finishedAt  timestamp,
    	error varchar(100)
		)`

	if _, err := p.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func NewPostgresStore(host string, user string, dbname string, password string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", host, user, dbname, password)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db:   db,
		name: "hello",
	}, nil
}

func (p PostgresStore) CreateTask(t *task.Task) error {
	query := `
		insert into tasks
		(id, priority, taskType, status, backOffDuration, payload, createdAt, createdBy, error)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		returning id
		`

	_, err := p.db.Exec(
		query,
		t.Id,
		t.Priority,
		t.TaskType,
		t.Status,
		t.BackOffDuration,
		t.Payload,
		t.CreatedAt,
		t.CreatedBy,
		t.ErrorDetails)

	if err != nil {
		slog.Error(err.Error())
		return err
	}
	return nil
}

func (p PostgresStore) UpdateTask(t *task.Task) error {

	// Prepare the SQL update statement
	sqlStatement := `
        UPDATE tasks
        SET status = $2, startedAt = $3, finishedAt = $4, error = $5
        WHERE id = $1;`

	// Execute the update statement
	res, err := p.db.Exec(sqlStatement, t.Id, t.Status, t.StartedAt, t.FinishedAt, t.ErrorDetails)
	if err != nil {
		log.Fatal(err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		slog.Error("error while writing task to database: ", err)
		return err
	}

	if count == 0 {
		return errors.New("failed to find and update task id in database")
	}

	return nil
}

func (p PostgresStore) GetTaskById(id string) (*task.Task, error) {
	rows, err := p.db.Query("select * from tasks where id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoTask(rows)
	}

	return nil, fmt.Errorf("task %s not found", id)
}

func scanIntoTask(rows *sql.Rows) (*task.Task, error) {
	t := new(task.Task)

	err := rows.Scan(
		&t.Id,
		&t.Priority,
		&t.TaskType,
		&t.Status,
		&t.BackOffDuration,
		&t.Payload,
		&t.CreatedAt,
		&t.CreatedBy,
		&t.StartedAt,
		&t.FinishedAt,
		&t.ErrorDetails)

	if err != nil {
		return nil, err
	}

	// Not ideal but I needed a quick workaround to save in case task has an error
	if t.ErrorDetails != "" {
		t.Error = errors.New(t.ErrorDetails)
	}

	return t, nil
}
