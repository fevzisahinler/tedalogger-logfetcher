// internal/logfetcher/elastic.go

package logfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"tedalogger-logfetcher/config"
)

func connectES() (*elasticsearch.Client, error) {
	cfg := config.GetConfig()

	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.ElasticURL},
	}

	if cfg.ElasticUser != "" && cfg.ElasticPass != "" {
		esConfig.Username = cfg.ElasticUser
		esConfig.Password = cfg.ElasticPass
	}

	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch info error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error response: %s", res.String())
	}

	log.Printf("Connected to Elasticsearch at %s (Auth? %v)", cfg.ElasticURL, cfg.ElasticUser != "")
	return es, nil
}

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
