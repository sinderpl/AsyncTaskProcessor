build:
	go build -o builds/bin/ main.go

run:
	go run main.go

test:
	@go test -v ./..

#db:
#	@docker run --name postgres -e POSTGRES_PASSWORD=asyncProcessor -p 5432:5432 -d postgres
#	@sleep 1 #Wait for container boot