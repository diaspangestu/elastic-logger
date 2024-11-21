package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
)

type Logger struct {
	client  *elasticsearch.Client
	index   string
	service string
}

type LogMessage struct {
	ServiceName string `json:"service_name"`
	Hostname    string `json:"hostname"`
	Timestamp   string `json:"timestamp"`
	HttpStatus  string `json:"http_status"`
	Header      string `json:"header"`
	Request     string `json:"request"`
	Response    string `json:"response"`
}

func New(elasticURL, index, service string) (*Logger, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{elasticURL},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elastic client: %w", err)
	}

	logger := &Logger{
		client:  client,
		index:   index,
		service: service,
	}

	return logger, nil
}

func (l *Logger) Log(logMsg LogMessage) error {
	data, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	res, err := l.client.Index(
		l.index,
		bytes.NewReader(data),
		l.client.Index.WithContext(context.Background()),
	)
	if err != nil {
		return fmt.Errorf("failed to send log to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch returned an error: %s", res.String())
	}

	return nil
}
