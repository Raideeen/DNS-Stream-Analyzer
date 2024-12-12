package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "grpc-server:50051" // Address of the gRPC server with the hostname defined in the compose.yml
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
	// Sleep a bit to let the server setup
	time.Sleep(10 * time.Second)

	// Connect to the gRPC server
	dialOption := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(address, dialOption)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDnsServiceClient(conn)

	// Kafka consumer test
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "broker:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	err = c.SubscribeTopics([]string{"myTopic", "^aRegex.*[Tt]opic"}, nil)

	if err != nil {
		panic(err)
	}

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

		// Kafka consumer test
		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else if !err.(kafka.Error).IsTimeout() {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}

		time.Sleep(100 * time.Millisecond)
	}

	c.Close()
}
