package queue_reader

import (
	"Monitor/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const PollIntervalMilliSeconds = 100

type QueueReader struct {
	ctx       context.Context
	wg        *sync.WaitGroup
	cfg       *config.Config
	mutex     *sync.RWMutex
	writeChan chan string
	checkChan chan string
}

func NewQueueReader(
	ctx context.Context,
	wg *sync.WaitGroup,
	cfg *config.Config,
	mutex *sync.RWMutex,
	writeChan chan string,
	checkChan chan string,
) *QueueReader {
	return &QueueReader{
		ctx:       ctx,
		wg:        wg,
		cfg:       cfg,
		mutex:     mutex,
		writeChan: writeChan,
		checkChan: checkChan,
	}
}

func (qr *QueueReader) Start() {
	go func() {
		qr.mutex.RLock()
		brokers := qr.cfg.Kafka.Brokers
		groupID := qr.cfg.Kafka.GroupID
		readTopic := qr.cfg.Kafka.ReadTopic
		qr.mutex.RUnlock()
		err := qr.consumeKafkaMessages(brokers, groupID, []string{readTopic})
		if err != nil {
			log.Fatal("Queue Reader failed to start ...")
			return
		}
	}()

	log.Printf("Queue Reader started ...")
}

func (qr *QueueReader) consumeKafkaMessages(brokers, groupID string, topics []string) error {
	conf := &kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": false,
	}

	consumer, err := kafka.NewConsumer(conf)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer func(consumer *kafka.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Fatal("Failed to close consumer: %v\n", err)
		}
	}(consumer)

	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	run := true
	for run {
		select {
		case sig := <-qr.ctx.Done():
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := consumer.Poll(PollIntervalMilliSeconds)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				qr.processKafkaMessages(string(e.Value))
				_, err := consumer.CommitMessage(e)
				if err != nil {
					log.Printf("Error committing message: %v\n", err)
				}

			case kafka.Error:
				log.Printf("Error: %v\n", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}

			default:
				log.Printf("Ignored event: %s\n", e)
			}
		}
	}

	return nil
}

func (qr *QueueReader) processKafkaMessages(message string) {
	log.Printf("Queue Reader: Kafka Message: %s\n", message)
	var result map[string]any
	if err := json.Unmarshal([]byte(message), &result); err != nil {
		panic(err)
	}
	log.Printf("Controller: %s", result["Controller"])
	// TODO: check Controller and Action names
	if result["Controller"] == "Data" && result["Action"] == "Copy" {
		qr.checkChan <- message
		return
	}
	qr.writeChan <- message
}
