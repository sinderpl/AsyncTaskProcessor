package storage

import (
	"database/sql"
	"fmt"
	"github.com/sinderpl/AsyncTaskProcessor/task"
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
		id string primary key,
		priority int,
    	taskType varchar(30),
    	status varchar(60),
    	backOffDuration int,
    	processableTask jsonb,      
        createdAt timestamp,
        createdBy string,
    	error string,
		)`

	if _, err := p.db.Exec(query); err != nil {
		return err
	}

	return nil
}

func NewPostgresStore(user string, dbname string, password string) (*PostgresStore, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable", user, dbname, password)
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
		(id, priority, taskType, status, backOffDuration, processableTask, createdAt, createdBy, error)
		values ($1, $2, $3, $4, $5)
		returning id
		`

	resp, err := p.db.Exec(
		query,
		t.Id,
		t.Priority,
		t.TaskType,
		t.Status,
		t.BackOffDuration,
		t.ProcessableTask,
		t.CreatedAt,
		t.CreatedBy,
		t.Error)

	fmt.Printf("sql create task result %v \n", resp)

	if err != nil {
		return err
	}
	return nil
}

func (p PostgresStore) UpdateTask(t *task.Task) error {
	//TODO implement me
	panic("implement me")
}

func (p PostgresStore) GetTaskById(id string) (*task.Task, error) {
	rows, err := p.db.Query("select * from tasks where id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoTask(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func scanIntoTask(rows *sql.Rows) (*task.Task, error) {
	t := new(task.Task)
	err := rows.Scan(
		&t.Id,
		&t.Priority,
		&t.TaskType,
		&t.Status,
		&t.BackOffDuration,
		&t.ProcessableTask,
		&t.CreatedAt,
		&t.CreatedBy,
		&t.Error)

	if err != nil {
		return nil, err
	}

	return t, nil
}
