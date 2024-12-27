#
# Docker service for testing
#

# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download the Go modules
RUN go mod download

# Copy the entire project
COPY . .

# Install the test dependencies
RUN go get -u github.com/stretchr/testify

# Run the tests
CMD ["go", "test", "./..."]