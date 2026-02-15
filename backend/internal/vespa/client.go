package vespa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yourusername/vespa-knowledge-hub/internal/models"
)

// Client represents a Vespa client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Vespa client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// IndexDocument indexes a document in Vespa
func (c *Client) IndexDocument(item *models.KnowledgeItem) error {
	docURL := fmt.Sprintf("%s/document/v1/default/knowledge_item/docid/%s",
		c.BaseURL, url.PathEscape(item.ID))

	vespaDoc := item.ToVespaDocument()
	jsonData, err := json.Marshal(vespaDoc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req, err := http.NewRequest("POST", docURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vespa returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteDocument deletes a document from Vespa
func (c *Client) DeleteDocument(docID string) error {
	docURL := fmt.Sprintf("%s/document/v1/default/knowledge_item/docid/%s",
		c.BaseURL, url.PathEscape(docID))

	req, err := http.NewRequest("DELETE", docURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vespa returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// HealthCheck checks if Vespa is healthy
func (c *Client) HealthCheck() error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/state/v1/health")
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("vespa is not healthy: status %d", resp.StatusCode)
	}

	return nil
}
