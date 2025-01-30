// internal/logfetcher/manager.go

package logfetcher

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type consumerHandle struct {
	cancelFunc context.CancelFunc
	queueName  string
}

func StartManager() {
	consumers := make(map[string]*consumerHandle)
	var mu sync.Mutex

	go func() {
		for {
			nasList, err := fetchNASList()
			if err != nil {
				log.Printf("Error fetching NAS list: %v", err)
				time.Sleep(10 * time.Second)
				continue
			}

			desiredQueues := make(map[string]bool)
			for _, nas := range nasList {
				queueName := fmt.Sprintf("%s-%s-queue",
					strings.ToLower(nas.Brand),
					nas.Nasname,
				)
				desiredQueues[queueName] = true
			}

			mu.Lock()
			for q, handle := range consumers {
				if !desiredQueues[q] {
					log.Printf("Stopping consumer for removed queue: %s", q)
					handle.cancelFunc()
					delete(consumers, q)
				}
			}

			for q := range desiredQueues {
				if _, exists := consumers[q]; !exists {
					log.Printf("Starting consumer for new queue: %s", q)

					ctx, cancel := context.WithCancel(context.Background())
					ch := &consumerHandle{
						cancelFunc: cancel,
						queueName:  q,
					}
					consumers[q] = ch

					go func(qName string) {
						brand, nasIP := parseQueueName(qName)
						if err := startConsumer(ctx, qName, brand, nasIP); err != nil {
							log.Printf("Consumer %s error: %v", qName, err)
						}
					}(q)
				}
			}
			mu.Unlock()

			time.Sleep(10 * time.Second)
		}
	}()

	select {}
}

func parseQueueName(qName string) (string, string) {
	parts := strings.Split(qName, "-")
	if len(parts) < 3 {
		return "unknown", "unknown"
	}
	return parts[0], parts[1]
}
