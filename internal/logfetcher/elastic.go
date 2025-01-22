package logfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

// getElasticURL returns the ELASTIC_URL from environment
func getElasticURL() string {
	return getEnv("ELASTIC_URL", "http://localhost:9200")
}

// getElasticUser returns ELASTIC_USER
func getElasticUser() string {
	return getEnv("ELASTIC_USER", "")
}

// getElasticPass returns ELASTIC_PASS
func getElasticPass() string {
	return getEnv("ELASTIC_PASS", "")
}

func connectES() (*elasticsearch.Client, error) {
	esURL := getElasticURL()
	esUser := getElasticUser()
	esPass := getElasticPass()

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}

	// Only set username/password if not empty
	if esUser != "" && esPass != "" {
		cfg.Username = esUser
		cfg.Password = esPass
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// quick test
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch info error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error response: %s", res.String())
	}

	log.Printf("Connected to Elasticsearch at %s (Auth? %v)", esURL, esUser != "")
	return es, nil
}

// indexLogToES indexes `doc` into `indexName`.
func indexLogToES(es *elasticsearch.Client, doc ParsedLog, indexName string) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	res, err := es.Index(
		indexName,
		strings.NewReader(string(data)),
		es.Index.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("ES index error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ES response error: %s", res.String())
	}
	return nil
}
