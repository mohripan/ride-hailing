package daprpubsub

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ride-hailing/shared/pkg/outbox"
)

type Publisher struct {
	baseURL string
	client  *http.Client
}

func NewPublisher(httpPort, pubsubName string) *Publisher {
	return &Publisher{
		baseURL: fmt.Sprintf("http://127.0.0.1:%s/v1.0/publish/%s", httpPort, url.PathEscape(pubsubName)),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, msg outbox.Message) error {
	query := url.Values{}
	query.Set("metadata.rawPayload", "true")
	query.Set("metadata.key", msg.Key)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s?%s", p.baseURL, url.PathEscape(msg.Topic), query.Encode()),
		bytes.NewReader(msg.Payload),
	)
	if err != nil {
		return fmt.Errorf("create dapr publish request for topic %q: %w", msg.Topic, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("publish topic %q via dapr: %w", msg.Topic, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("publish topic %q via dapr returned %s: %s", msg.Topic, resp.Status, strings.TrimSpace(string(body)))
	}

	return nil
}
