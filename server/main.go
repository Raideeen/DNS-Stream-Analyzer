package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port      string = ":50051"
	redisAddr string = "redis:6379"
)

var topic string = "myTopic"

type server struct {
	pb.UnimplementedDnsServiceServer
	redisClient *redis.Client
	producer    *kafka.Producer
}

// SendDnsRequest handles incoming DNS requests
func (s *server) SendDnsRequest(ctx context.Context, req *pb.DnsRequest) (*pb.DnsResponse, error) {
	// Check if IP is already marked as malicious in Redis
	val, err := s.redisClient.Get(ctx, req.GetIpAddress()).Result()
	if err == nil && val == "malicious" {
		log.Printf("Blacklisted IP detected, blocking: %s", req.GetIpAddress())
		return &pb.DnsResponse{Status: "blocked"}, nil
	}

	// Produce message to Kafka topic
	message := fmt.Sprintf("IP: %s, Domain: %s, QueryType: %s, Timestamp: %d",
		req.GetIpAddress(), req.GetDomain(), req.GetQueryType(), req.GetTimestamp())

	s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	log.Printf("Sent DNS request to Kafka: %v", message)
	return &pb.DnsResponse{Status: "success"}, nil
}

// BlockIp handles blocking IPs based on consumer feedback
func (s *server) BlockIp(ctx context.Context, req *pb.BlockIpRequest) (*pb.BlockIpResponse, error) {
	err := s.redisClient.Set(ctx, req.GetIpAddress(), "malicious", 0).Err()
	if err != nil {
		log.Printf("Failed to block IP: %v", err)
		return &pb.BlockIpResponse{Status: "failed"}, err
	}
	log.Printf("Blocked IP: %s", req.GetIpAddress())
	return &pb.BlockIpResponse{Status: "success"}, nil
}

func main() {
	// Redis setup
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	// Kafka create topic
	// Create admin client and topic before serving gRPC
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": "broker:9092"})
	if err != nil {
		log.Fatalf("Failed to create Kafka admin client: %v", err)
	}
	defer adminClient.Close()

	// Check and create the topic if it doesn't exist
	metadata, err := adminClient.GetMetadata(&topic, false, 5000)
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}
	if _, ok := metadata.Topics[topic]; !ok {
		results, err := adminClient.CreateTopics(
			context.Background(),
			[]kafka.TopicSpecification{{Topic: topic, NumPartitions: 3, ReplicationFactor: 2}},
			nil,
		)
		if err != nil {
			log.Fatalf("Failed to create topic: %v", err)
		}
		for _, result := range results {
			if result.Error.Code() != kafka.ErrNoError {
				log.Fatalf("Failed to create topic %s: %s", result.Topic, result.Error.String())
			}
		}
		log.Printf("Topic %s created successfully", topic)
	} else {
		log.Printf("Topic %s already exists", topic)
	}

	// Kafka producer setup
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "broker:9092"})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Start producer delivery report handler in a separate goroutine
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Start gRPC server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterDnsServiceServer(grpcServer, &server{
		redisClient: redisClient,
		producer:    producer,
	})

	log.Printf("Server is listening on %v", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
