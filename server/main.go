package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port     string = ":50051"
	mongoURI string = "mongodb://db:27017"
	dbName   string = "dns_data"
	collName string = "requests"
)

var topic string = "myTopic"

type server struct {
	pb.UnimplementedDnsServiceServer
	mongoClient *mongo.Client
	producer    *kafka.Producer
}

// SendDnsRequest handles incoming DNS requests
func (s *server) SendDnsRequest(ctx context.Context, req *pb.DnsRequest) (*pb.DnsResponse, error) {
	collection := s.mongoClient.Database(dbName).Collection(collName)

	// Insert request into MongoDB
	document := bson.D{
		{Key: "ip_address", Value: req.GetIpAddress()},
		{Key: "domain", Value: req.GetDomain()},
		{Key: "query_type", Value: req.GetQueryType()},
		{Key: "timestamp", Value: req.GetTimestamp()},
	}

	_, err := collection.InsertOne(ctx, document)
	if err != nil {
		log.Printf("Failed to insert document: %v", err)
		return &pb.DnsResponse{Status: "failed"}, err
	}

	// Produce message to Kafka topic
	message := fmt.Sprintf("IP: %s, Domain: %s, QueryType: %s, Timestamp: %d",
		req.GetIpAddress(), req.GetDomain(), req.GetQueryType(), req.GetTimestamp())

	s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	log.Printf("Inserted DNS request and sent to Kafka: %v", document)
	return &pb.DnsResponse{Status: "success"}, nil
}

func main() {
	// MongoDB setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB!")

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
			[]kafka.TopicSpecification{{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}},
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
		mongoClient: mongoClient,
		producer:    producer,
	})

	log.Printf("Server is listening on %v", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
