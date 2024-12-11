package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port     = ":50051"
	mongoURI = "mongodb://db:27017" // "db" is the hostname of the mongoDB image in the compose.yml
	dbName   = "dns_data"
	collName = "requests"
)

// server implements the gRPC service
type server struct {
	pb.UnimplementedDnsServiceServer

	mongoClient *mongo.Client
}

// SendDnsRequest handles incoming DNS requests
func (s *server) SendDnsRequest(
	ctx context.Context, req *pb.DnsRequest,
) (*pb.DnsResponse, error) {
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

	log.Printf("Inserted DNS request: %v", document)
	return &pb.DnsResponse{Status: "success"}, nil
}

func main() {
	// Context for MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB connection setup
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("Disconnected from MongoDB")
		}
	}()

	// Test MongoDB connection
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	log.Println("Connected to MongoDB!")

	// Start gRPC server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)

	pb.RegisterDnsServiceServer(grpcServer, &server{mongoClient: mongoClient})

	log.Printf("Server is listening on %v", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
