package logfetcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

func startConsumer(ctx context.Context, queueName, brand, nasIP string) error {
	amqpURL := getRabbitMQURL()
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("RabbitMQ dial error: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("Channel error: %w", err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("Consume error: %w", err)
	}

	esClient, err := connectES()
	if err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("Elasticsearch connection error: %w", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}

				rawMsg := string(d.Body)

				lm := LogMessage{
					Message: rawMsg,
				}

				doc := parseAndDetermineBrand(lm)

				doc.NASName = nasIP

				if doc.URL == "" {
					continue
				}
				if doc.Brand != "forti" && doc.Brand != "ruijie" {
					continue
				}

				indexName := fmt.Sprintf("%s-%s",
					strings.ReplaceAll(nasIP, ".", "_"),
					dateString(),
				)

				if err := indexLogToES(esClient, doc, indexName); err != nil {
					log.Printf("ES index error (queue=%s): %v", queueName, err)
				} else {
					log.Printf("Indexed (queue=%s) brand=%s, url=%s, index=%s",
						queueName, doc.Brand, doc.URL, indexName)
				}
			}
		}
	}()

	<-done
	ch.Close()
	conn.Close()
	return nil
}

func getRabbitMQURL() string {
	url := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672")
	return url
}

func dateString() string {
	return now().Format("02-01-2006")
}

var now = func() (t time.Time) {
	return time.Now()
}
