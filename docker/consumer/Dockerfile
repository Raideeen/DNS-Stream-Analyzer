#
# Compiling consumer application (pull from Kafka)
#

# Build the consumer application
FROM golang:latest AS consumer-build 

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download 

# Copy the source code 
COPY consumer/ consumer/
COPY proto/ proto/
COPY pb/ pb/

RUN GOOS=linux go build -o /consumer-app ./consumer/main.go

# Create a minimal image for the consumer application
FROM gcr.io/distroless/base-debian12 AS consumer
COPY --from=consumer-build /consumer-app /consumer-app
EXPOSE 50051
USER nonroot:nonroot
ENTRYPOINT [ "/consumer-app" ]