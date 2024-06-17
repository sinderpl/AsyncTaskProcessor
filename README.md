# AsyncTaskProcessor

### Overview
An asynchronous task processing system that should be easily extendable to accommodate new type of tasks. <br/>
The system utilises a priority based queue system with backoff, retries and persistence for task information.

### Running
Docker:
```
docker-compose up -d
```
Go locally (requires go installed):
```
make db
make up
```

All of the requests and postman collection can be found in api/requests to easily import and test. <br/>
# Endpoints:
#### GET /healthz - simple endpoint to check if service is up and running
```
curl --location 'http://localhost:8080/healthz'
```

#### GET /task/{id} - simple endpoint to check if service is up and running
```
curl --location 'http://localhost:8080/task/{taskId}'
```

#### POST /tasks/enqueue - simple endpoint to check if service is up and running
```
curl --location 'http://localhost:8080/tasks/enqueue' \
--header 'Content-Type: application/json' \
--data-raw '{
  "tasks" : [
    {
      "taskType" : "SendEmail",
      "payload": {
        "sendTo" : ["helloworld@test.com", "helloWorld2@test.com"],
        "sendFrom": "hello1@test.com",
        "subject" : "Hi !",
        "body" : "hope you are well"
      }
    },
    {
      "taskType" : "GenerateReport",
      "priority" : 1,
      "payload" : {
        "notify" : ["helloworld@test.com", "helloWorld2@test.com"],
        "reportType" : "Financial Report"
      }
    },
    {
      "taskType" : "CPUProcess",
      "priority" : 1,
      "payload" : {
        "ProcessType" : "want to fail"
      },
      "backOffDuration" : "5s"
    }
  ]
}'
```

#### POST /task/{taskId}/retry - simple endpoint to check if service is up and running
```
curl --location --request POST 'http://localhost:8080/task/e83a5116-0191-462c-8cf7-18c21a3a4939/retry' \
--data ''
```

### Architecture


### Design Considerations


### Configurations
```
api:
  listenAddr: ':8080'
queue:
  maxBufferSize: 10 
  workerPoolSize: 5
  maxTaskRetry: 3 
storage:
  host: 'db'
  user: 'postgres'
  dbname: 'postgres'
  password: 'asyncProcessor'
```

Outside of the api address and storage info we can configure the queue to our expectations
```
# sets the max size of the buffer channels which means how many tasks can wait in a channel to be processed
# tasks outside of the buffer will still be tracked in the queue this is to control the channel size 
maxBufferSize: 10
```
```
workerPoolSize: 5 # amount of workers processing tasks
```

```
maxTaskRetry: 3 # max retry on tasks when they fail
```



### Further work to consider:
- [ ] Update config file to match dockerfile and be read from one place 
- [ ] Tests
- [ ] Queue Management - Being able to prioritise queues dynamically and add more queues when needed
- [ ] Queue prioritisation (avoid starvation for low priority tasks by making sure they are executed from time to time)
- [ ] Batch task creation for DB
- [ ] Improve architecture diagram
- [ ] 

## TODO
- [x] Config
- [x] API
- [x] Endpoint for enqueue tasks
- [x] Endpoint for querying tasks
- [x] Retry failed endpoint
- [x] Generic Task
- [ ] Queue Management
- [ ] Queue prioritisation (avoid starvation)
- [x] Storage database
- [x] Add docker-compose db to create persistence through sql
- [x] Async Processing
- [x] Error handling, retries
- [x] Dead Letter Queue
- [x] Logging 
- [ ] Tests
- [x] Code Documentation
- [x] System Documentation

Bonus
- [x] Max concurrent limit
- [x] Retries with backoff
- [x] Persistence


Final
- [x] Lint and format imports
 