
build:
	go build -o builds/bin/ main.go

run:
	go run main.go

test:
	go test ./pkg/...