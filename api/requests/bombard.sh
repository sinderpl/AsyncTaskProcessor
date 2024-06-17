#!/bin/bash

# Used to load test the api, go install required https://github.com/codesenberg/bombardier

bombardier -c 3 -n 20 -d 6s \
-H 'Content-Type: application/json' \
-H 'x-api-key:test' -m POST -f request-body.json -l http://localhost:8080/tasks/enqueue