# DNS-Stream-Analyzer

Cloud based project designed to simulate DNS traffic detection to detect malicious activity. Using Kafka for data distribution and gRPC for the back-end.

## Quick start

This project has been done on Ubuntu 22.04.5 LTS (Jammy Jellyfish) on WSL2.

```bash
# First update your system based on your distro
sudo apt update && sudo apt upgrade 

# Install MongoDB for Ubuntu 22.04.05 LTS (Jammy Jellyfish)
sudo apt-get install gnupg curl
curl -fsSL https://www.mongodb.org/static/pgp/server-8.0.asc | \
   sudo gpg -o /usr/share/keyrings/mongodb-server-8.0.gpg \
   --dearmor
echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-8.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/8.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-8.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org

# Install Go
wget https://go.dev/dl/go1.23.3.linux-amd64.tar.gz -P /tmp

sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go1.23.3.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin
go version # go version go1.23.3 linux/amd64

# Install protobuf
sudo apt install protobuf-compiler

# Install gRPC go module
go get go.mongodb.org/mongo-driver/v2/mongo
go get google.golang.org/grpc 
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest # Allow for interact with the server easily using server reflection
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 
# Add the go modules in path
export PATH=$PATH:$(go env GOPATH)/bin
source ~/.bashrc

# Generate the protobuf files 
make generate
```
