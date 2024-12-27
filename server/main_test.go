package main

import (
	"context"
	"testing"
	"time"

	"github.com/Raideeen/DNS-Stream-Analyzer/pb"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	testAddress   = "grpc-server:50051"
	testRedisAddr = "redis:6379"
)

func TestBlacklistIP(t *testing.T) {
	// Setup Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: testRedisAddr,
	})
	defer redisClient.Close()

	// Flush the Redis database
	err := redisClient.FlushDB(redisClient.Context()).Err()
	if err != nil {
		t.Fatalf("Failed to flush Redis database: %v", err)
	}

	// Connect to the gRPC server
	conn, err := grpc.NewClient(testAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewDnsServiceClient(conn)

	// Define the test IP address
	testIPAddress := "192.168.1.70"

	// First connection should succeed
	req := &pb.DnsRequest{
		IpAddress: testIPAddress,
		Domain:    "test.com",
		QueryType: "A",
		Timestamp: time.Now().Unix(),
	}
	resp, err := client.SendDnsRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp.GetStatus())

	// Block the IP address
	blockReq := &pb.BlockIpRequest{
		IpAddress: testIPAddress,
	}
	blockResp, err := client.BlockIp(context.Background(), blockReq)
	assert.NoError(t, err)
	assert.Equal(t, "success", blockResp.GetStatus())

	// Second connection should be blocked
	resp, err = client.SendDnsRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "blocked", resp.GetStatus())

	// Third connection should also be blocked
	resp, err = client.SendDnsRequest(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "blocked", resp.GetStatus())
}
