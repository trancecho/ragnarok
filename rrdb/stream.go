package rrdb

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamClient provides a simple interface for Redis stream operations.
type StreamClient struct {
	rdb *redis.Client
}

// NewStreamClient creates a new client for Redis stream operations.
func NewStreamClient(rdb *redis.Client) *StreamClient {
	return &StreamClient{rdb: rdb}
}

// Publish sends an event to the specified stream.
// `values` is a map of key-value pairs that make up the event.
func (c *StreamClient) Publish(ctx context.Context, stream string, values map[string]interface{}) (string, error) {
	msgId, err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
	if err != nil {
		log.Printf("Error publishing to stream %s: %v", stream, err)
		return "", err
	}
	return msgId, nil
}

// MessageHandler is a function that processes a single message from the stream.
// It should return an error if the message processing fails, to prevent acknowledgment.
type MessageHandler func(message map[string]interface{}) error

// Subscribe sets up a consumer group and starts listening for new messages on a stream.
// This function will block and run indefinitely.
// `stream`: The name of the Redis stream.
// `group`: The consumer group name.
// `consumer`: A unique name for this consumer within the group.
// `handler`: The callback function to process each message.
func (c *StreamClient) Subscribe(ctx context.Context, stream, group, consumer string, handler MessageHandler) error {
	// Create the consumer group. If it already exists, this will do nothing.
	// We use MKSTREAM to create the stream if it doesn't exist.
	err := c.rdb.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Printf("Error creating consumer group '%s' for stream '%s': %v", group, stream, err)
		return err
	}

	log.Printf("Consumer '%s' starting to listen on stream '%s' with group '%s'", consumer, stream, group)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Consumer '%s' stopping.", consumer)
			return ctx.Err()
		default:
			// Read one message from the stream, block for 2 seconds if there are no new messages.
			streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: consumer,
				Streams:  []string{stream, ">"}, // ">" means only new messages
				Count:    1,
				Block:    2 * time.Second,
				NoAck:    false, // We will manually acknowledge messages
			}).Result()

			if err != nil {
				if errors.Is(err, redis.Nil) {
					// Timed out, just continue
					continue
				}
				log.Printf("Error reading from stream '%s': %v", stream, err)
				time.Sleep(1 * time.Second) // Wait a bit before retrying on error
				continue
			}

			// Process the message
			for _, s := range streams {
				for _, msg := range s.Messages {
					log.Printf("Consumer '%s' received message %s", consumer, msg.ID)

					processingErr := handler(msg.Values)
					if processingErr != nil {
						log.Printf("Error processing message %s: %v. Message will not be acknowledged.", msg.ID, processingErr)
						continue
					}

					// Acknowledge the message
					ackErr := c.rdb.XAck(ctx, stream, group, msg.ID).Err()
					if ackErr != nil {
						log.Printf("Error acknowledging message %s: %v", msg.ID, ackErr)
					}
				}
			}
		}
	}
}
