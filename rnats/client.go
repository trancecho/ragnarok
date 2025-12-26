package rnats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps NATS connection and JetStream
type Client struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// Config for NATS client
type Config struct {
	URL      string // NATS server URL, e.g., "nats://localhost:4222"
	Username string // Optional authentication username
	Password string // Optional authentication password
	Name     string // Connection name
}

// NewClient creates a new NATS client
func NewClient(cfg Config) (*Client, error) {
	opts := []nats.Option{
		nats.Name(cfg.Name),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1), // infinite reconnects
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				log.Printf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected")
		}),
	}

	// Authentication
	if cfg.Username != "" && cfg.Password != "" {
		opts = append(opts, nats.UserInfo(cfg.Username, cfg.Password))
	}

	// Connect to NATS
	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	log.Printf("NATS client connected to %s", cfg.URL)

	return &Client{
		conn: conn,
		js:   js,
	}, nil
}

// Close closes the NATS connection
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		log.Println("NATS client closed")
	}
}

// Connection returns the underlying NATS connection (for advanced usage)
func (c *Client) Connection() *nats.Conn {
	return c.conn
}

// JetStream returns the JetStream context (for advanced usage)
func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}

// Ping checks if the connection is alive
func (c *Client) Ping(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		err := c.conn.FlushWithContext(ctx)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// ============ Publish-Subscribe Methods ============

// MessageHandler is a function that processes a single message.
type MessageHandler func(data []byte) error

// Publish sends a message to the specified subject.
// The data will be automatically serialized to JSON.
func (c *Client) Publish(ctx context.Context, subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = c.js.Publish(ctx, subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish to subject %s: %w", subject, err)
	}
	return nil
}

// PublishRaw sends raw bytes to the specified subject without JSON serialization.
func (c *Client) PublishRaw(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish to subject %s: %w", subject, err)
	}
	return nil
}

// Subscribe sets up a durable consumer and starts listening for messages on a subject.
// This function will block and run indefinitely.
func (c *Client) Subscribe(ctx context.Context, subject, stream, consumer string, handler MessageHandler) error {
	// Get or create stream
	streamObj, err := c.getOrCreateStream(ctx, stream, subject)
	if err != nil {
		return fmt.Errorf("failed to get/create stream: %w", err)
	}

	// Create consumer
	cons, err := streamObj.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       consumer,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: 100,
		AckWait:       30 * time.Second,
		FilterSubject: subject,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	log.Printf("Consumer '%s' starting to listen on subject '%s' in stream '%s'", consumer, subject, stream)

	// Consume messages
	_, err = cons.Consume(func(msg jetstream.Msg) {
		processingErr := handler(msg.Data())
		if processingErr != nil {
			log.Printf("Error processing message: %v. Message will not be acknowledged.", processingErr)
			if nackErr := msg.Nak(); nackErr != nil {
				log.Printf("Error nacking message: %v", nackErr)
			}
			return
		}

		if ackErr := msg.Ack(); ackErr != nil {
			log.Printf("Error acknowledging message: %v", ackErr)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	// Block until context is cancelled
	<-ctx.Done()
	log.Printf("Consumer '%s' stopping.", consumer)
	return ctx.Err()
}

// SubscribeWithJSON is similar to Subscribe but automatically deserializes JSON messages.
func (c *Client) SubscribeWithJSON(ctx context.Context, subject, stream, consumer string, handler func(data map[string]interface{}) error) error {
	return c.Subscribe(ctx, subject, stream, consumer, func(data []byte) error {
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return handler(msg)
	})
}

// Request sends a request and waits for a reply (request-reply pattern)
func (c *Client) Request(ctx context.Context, subject string, data interface{}, timeout time.Duration) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	msg, err := c.conn.RequestWithContext(ctx, subject, payload)
	if err != nil {
		if errors.Is(err, nats.ErrTimeout) {
			return nil, fmt.Errorf("request timeout after %v", timeout)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return msg.Data, nil
}

// ============ KeyValue Methods ============

// KVPutWithTTL stores a value in the KV store with TTL.
// The value will be automatically serialized to JSON.
// TTL is specified as time.Duration (e.g., 24*time.Hour)
func (c *Client) KVPutWithTTL(ctx context.Context, bucket, key string, value interface{}, ttl time.Duration) error {
	kv, err := c.getOrCreateKVWithTTL(ctx, bucket, ttl)
	if err != nil {
		return err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	_, err = kv.Put(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}
	return nil
}

// KVPutRaw stores raw bytes in the KV store without JSON serialization.
func (c *Client) KVPutRaw(ctx context.Context, bucket, key string, value []byte) error {
	kv, err := c.getOrCreateKV(ctx, bucket)
	if err != nil {
		return err
	}

	_, err = kv.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}
	return nil
}

// KVGet retrieves a value from the KV store.
// The value will be automatically deserialized from JSON into the target.
// Returns an error if the bucket or key doesn't exist.
func (c *Client) KVGet(ctx context.Context, bucket, key string, target interface{}) error {
	kv, err := c.getKV(ctx, bucket)
	if err != nil {
		return err
	}

	entry, err := kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(entry.Value(), target); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return nil
}

// KVGetRaw retrieves raw bytes from the KV store without JSON deserialization.
// Returns an error if the bucket or key doesn't exist.
func (c *Client) KVGetRaw(ctx context.Context, bucket, key string) ([]byte, error) {
	kv, err := c.getKV(ctx, bucket)
	if err != nil {
		return nil, err
	}

	entry, err := kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return entry.Value(), nil
}

// KVDelete removes a key from the KV store.
func (c *Client) KVDelete(ctx context.Context, bucket, key string) error {
	kv, err := c.getOrCreateKV(ctx, bucket)
	if err != nil {
		return err
	}

	err = kv.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// KVExists checks if a key exists in the KV store.
func (c *Client) KVExists(ctx context.Context, bucket, key string) (bool, error) {
	kv, err := c.getOrCreateKV(ctx, bucket)
	if err != nil {
		return false, err
	}

	_, err = kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return true, nil
}

// KVKeys returns all keys in the bucket.
func (c *Client) KVKeys(ctx context.Context, bucket string) ([]string, error) {
	kv, err := c.getOrCreateKV(ctx, bucket)
	if err != nil {
		return nil, err
	}

	keys, err := kv.Keys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	return keys, nil
}

// KVWatch watches for changes to a key or key pattern in a bucket.
func (c *Client) KVWatch(ctx context.Context, bucket, keyPattern string, handler func(key string, value []byte, deleted bool)) error {
	kv, err := c.getOrCreateKV(ctx, bucket)
	if err != nil {
		return err
	}

	watcher, err := kv.Watch(ctx, keyPattern)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	go func() {
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case entry := <-watcher.Updates():
				if entry == nil {
					return
				}

				key := entry.Key()
				value := entry.Value()
				deleted := entry.Operation() == jetstream.KeyValueDelete

				handler(key, value, deleted)
			}
		}
	}()

	return nil
}

// ============ Helper Methods ============

// getOrCreateStream gets or creates a JetStream stream
func (c *Client) getOrCreateStream(ctx context.Context, streamName, subject string) (jetstream.Stream, error) {
	stream, err := c.js.Stream(ctx, streamName)
	if err == nil {
		return stream, nil
	}

	if !errors.Is(err, jetstream.ErrStreamNotFound) {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}

	stream, err = c.js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
		Storage:  jetstream.FileStorage,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	log.Printf("Created new stream '%s' with subject '%s'", streamName, subject)
	return stream, nil
}

// getKV gets an existing KV bucket (does not create)
func (c *Client) getKV(ctx context.Context, bucket string) (jetstream.KeyValue, error) {
	kv, err := c.js.KeyValue(ctx, bucket)
	if err != nil {
		if errors.Is(err, jetstream.ErrBucketNotFound) {
			return nil, fmt.Errorf("KV bucket not found: %s", bucket)
		}
		return nil, fmt.Errorf("failed to get KV bucket: %w", err)
	}
	return kv, nil
}

// getOrCreateKV gets or creates a KeyValue store (without TTL)
func (c *Client) getOrCreateKV(ctx context.Context, bucket string) (jetstream.KeyValue, error) {
	kv, err := c.js.KeyValue(ctx, bucket)
	if err == nil {
		return kv, nil
	}

	if !errors.Is(err, jetstream.ErrBucketNotFound) {
		return nil, fmt.Errorf("failed to get KV bucket: %w", err)
	}

	kv, err = c.js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: bucket,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create KV bucket: %w", err)
	}

	log.Printf("Created new KV bucket '%s'", bucket)
	return kv, nil
}

// getOrCreateKVWithTTL gets or creates a KeyValue store with TTL
func (c *Client) getOrCreateKVWithTTL(ctx context.Context, bucket string, ttl time.Duration) (jetstream.KeyValue, error) {
	kv, err := c.js.KeyValue(ctx, bucket)
	if err == nil {
		return kv, nil
	}

	if !errors.Is(err, jetstream.ErrBucketNotFound) {
		return nil, fmt.Errorf("failed to get KV bucket: %w", err)
	}

	kv, err = c.js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: bucket,
		TTL:    ttl, // Set TTL for all keys in this bucket
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create KV bucket: %w", err)
	}

	log.Printf("Created new KV bucket '%s' with TTL %v", bucket, ttl)
	return kv, nil
}
