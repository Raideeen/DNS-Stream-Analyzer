#
# Compiling client application
#

# Build the client application
FROM golang:latest AS client-build 

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download 

# Copy the source code 
COPY client/ client/
COPY proto/ proto/
COPY pb/ pb/

RUN GOOS=linux go build -o /client-app ./client/main.go

# Create a minimal image for the client application
FROM gcr.io/distroless/base-debian12 AS client
COPY --from=client-build /client-app /client-app
EXPOSE 50051
USER nonroot:nonroot
ENTRYPOINT [ "/client-app" ]