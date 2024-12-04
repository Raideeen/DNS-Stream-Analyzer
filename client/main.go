package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051" // Address of the gRPC server
)

var possibleDomains [4]string = [4]string{"mywebsite.com", "api.mywebsite.com", "cdn.mywebsite.com", "blog.mywebsite.com"}
var possibleQueryType [2]string = [2]string{"A", "AAAA"}

func randomIPAddress() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func randomDomain() string {
	return possibleDomains[rand.Intn(len(possibleDomains))]
}

func randomQueryType() string {
	return possibleQueryType[rand.Intn(len(possibleQueryType))]
}

func main() {
	// Connect to the gRPC server
	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(address, dialOption)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDnsServiceClient(conn)

	for {
		// Create a DNS request message

		req := &pb.DnsRequest{
			IpAddress: randomIPAddress(),
			Domain:    randomDomain(),
			QueryType: randomQueryType(),
			Timestamp: time.Now().Unix(),
		}

		// Send the request to the server
		_, err := client.SendDnsRequest(context.Background(), req)
		if err != nil {
			log.Printf("Error sending DNS request: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
