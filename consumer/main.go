package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "grpc-server:50051" // Address of the gRPC server
)

var topic string = "myTopic"

func extractIP(message string) string {
	// Extract IP from the message
	// Assuming the message format is "IP: <ip>, Domain: <domain>, QueryType: <query_type>, Timestamp: <timestamp>"
	parts := strings.Split(message, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "IP: ") {
			return strings.TrimSpace(strings.TrimPrefix(part, "IP: "))
		}
	}
	return ""
}

func isMalicious(ip string) bool {
	return strings.HasSuffix(ip, "70")
}

func main() {
	// Sleep a bit to let the server setup
	time.Sleep(10 * time.Second)

	// Kafka consumer setup
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "broker:9092",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	err = c.SubscribeTopics([]string{topic, "^aRegex.*[Tt]opic"}, nil)

	if err != nil {
		panic(err)
	}

	// Connect to the gRPC server
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewDnsServiceClient(conn)

	for {
		// Kafka consumer test
		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))

			// Analyze the message
			ip := extractIP(string(msg.Value))
			if isMalicious(ip) {
				// Send block request to the server
				req := &pb.BlockIpRequest{IpAddress: ip}
				_, err := client.BlockIp(context.Background(), req)
				if err != nil {
					log.Printf("Failed to send block IP request: %v", err)
				} else {
					log.Printf("Sent block IP request for IP: %s", ip)
				}
			}
		} else if !err.(kafka.Error).IsTimeout() {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
