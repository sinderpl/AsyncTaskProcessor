# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.12-alpine base image
FROM golang:1.12-alpine

# Setup
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build and run
RUN go build -o main .

# Expose port 8080 to the outside world
#EXPOSE 8080

# Run the executable
CMD ["./main"]