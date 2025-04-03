# Use the official Alpine-based Golang image to create a build artifact
FROM golang:1.24.1-alpine AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Install necessary dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Use the same Alpine image to package the compiled binary
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy everything from the previous stage
COPY --from=build /app/ .

# Command to run the executable
CMD ["./main"]
