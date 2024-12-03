package main

import (
	"context"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	// Create a new Kafka Producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092", // Address of the Kafka broker
		"client.id":         "myProducer",     // Unique identifier for this producer
	})
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	// Create a new Kafka Consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092", // Address of the Kafka broker
		"group.id":          "myGroup",        // Identifier for the consumer group
		"auto.offset.reset": "earliest",       // Start consuming from the beginning of the topic
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	// Create an AdminClient from the Producer (for topic management)
	adminClient, err := kafka.NewAdminClientFromProducer(producer)
	if err != nil {
		fmt.Printf("Failed to create admin client: %s\n", err)
		os.Exit(1)

	}
	defer adminClient.Close() // Close the AdminClient when done

	// Define the topic name
	topic := "Cloud_Data_Structure-Topic"

	topics, err := adminClient.GetMetadata(&topic, true, 5000)
	if err != nil {
		fmt.Printf("Failed to get metadata: %s\n", err)
		os.Exit(1)
	}

	// Check if the topic already exists
	if _, ok := topics.Topics[topic]; ok {
		fmt.Printf("Topic %s already exists\n", topic)
	} else {
		fmt.Printf("Topic %s does not exist, creating it...\n", topic)
		// Create the topic (if it doesn't exist)
		results, err := adminClient.CreateTopics(
			context.Background(),
			[]kafka.TopicSpecification{{ // Define a single topic specification
				Topic:             topic,
				NumPartitions:     1, // One partition for the topic
				ReplicationFactor: 1, // Replicate data to one broker (for simplicity)
			}},
			nil, // No additional options
		)
		if err != nil {
			fmt.Printf("Failed to create topic: %s\n", err)
			os.Exit(1)
		}

		// Check for errors during topic creation
		for _, result := range results {
			if result.Error.Code() != kafka.ErrNoError {
				fmt.Printf("Failed to create topic %s: %s\n", result.Topic, result.Error.String())
				os.Exit(1)
			}
		}
	}

	// Send a message to the topic
	producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte("Hello Go!"), // Message content
	}, nil)

	// Subscribe the consumer to the topic
	consumer.Subscribe(topic, nil) // Subscribe to receive messages from this topic

	for {
		msg, err := consumer.ReadMessage(-1) // Read a message with a timeout of -1 (indefinitely)
		if err == nil {
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else {
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)

		}
	}
}
