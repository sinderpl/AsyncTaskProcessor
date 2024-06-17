# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.18-alpine base image
#FROM golang:1.18-alpine as builder
#
## Install git
#RUN apt-get update && \
#    apt-get upgrade -y && \
#    apt-get install -y git
#
## Setup
#WORKDIR /app
#COPY go.mod ./
#RUN go mod download
#COPY . .
#
## Build and run
#RUN go build -o main .
#
## Expose port 8080 to the outside world
##EXPOSE 8080
#
## Run the executable
#CMD ["./main"]

# Use the official Golang image to build the application
FROM golang:1.22 as builder

#RUN apt update && apt install libc6

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Start a new stage from scratch
FROM golang:1.22

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

COPY config ./config

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
