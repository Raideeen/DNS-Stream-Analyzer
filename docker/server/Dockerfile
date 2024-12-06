#
# Compiling server application
#

# Build the server application
FROM golang:latest AS server-build

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download 

# Copy the source code 
COPY server/ server/
COPY proto/ proto/
COPY pb/ pb/

RUN CGO_ENABLED=0 GOOS=linux go build -o /server-app ./server/main.go

# Create a minimal image for the server application
FROM gcr.io/distroless/base-debian12 AS server
COPY --from=server-build /server-app /server-app
EXPOSE 50051
USER nonroot:nonroot
ENTRYPOINT [ "/server-app" ] 