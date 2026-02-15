package vespa

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yourusername/vespa-knowledge-hub/internal/models"
)

// VespaSearchResponse represents the response from Vespa search API
type VespaSearchResponse struct {
	Root struct {
		Fields struct {
			TotalCount int `json:"totalCount"`
		} `json:"fields"`
		Children []struct {
			ID        string                 `json:"id"`
			Relevance float64                `json:"relevance"`
			Fields    map[string]interface{} `json:"fields"`
		} `json:"children"`
	} `json:"root"`
}

// Search performs a search query against Vespa
func (c *Client) Search(query *models.SearchQuery) (*models.SearchResult, error) {
	yql := c.buildYQL(query)

	searchURL := fmt.Sprintf("%s/search/?yql=%s&hits=%d&offset=%d&query=%s",
		c.BaseURL,
		url.QueryEscape(yql),
		query.Limit,
		query.Offset,
		url.QueryEscape(query.Query),
	)

	resp, err := c.HTTPClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vespa returned status %d: %s", resp.StatusCode, string(body))
	}

	var vespaResp VespaSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&vespaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := &models.SearchResult{
		TotalCount: vespaResp.Root.Fields.TotalCount,
		Hits:       make([]models.SearchResultHit, 0, len(vespaResp.Root.Children)),
	}

	for _, child := range vespaResp.Root.Children {
		hit := models.SearchResultHit{
			ID:        child.ID,
			Relevance: child.Relevance,
			Fields:    child.Fields,
		}
		result.Hits = append(result.Hits, hit)
	}

	return result, nil
}

// buildYQL builds a YQL query from SearchQuery
func (c *Client) buildYQL(query *models.SearchQuery) string {
	var conditions []string

	// Text search using userQuery
	if query.Query != "" {
		conditions = append(conditions, "userQuery()")
	}

	// Filters
	if query.SourceType != "" {
		conditions = append(conditions, fmt.Sprintf("source_type contains '%s'", query.SourceType))
	}
	if query.RepoName != "" {
		conditions = append(conditions, fmt.Sprintf("repo_name contains '%s'", query.RepoName))
	}
	if query.Language != "" {
		conditions = append(conditions, fmt.Sprintf("language contains '%s'", query.Language))
	}

	whereClause := "true"
	if len(conditions) > 0 {
		whereClause = strings.Join(conditions, " and ")
	}

	yql := fmt.Sprintf("select * from knowledge_item where %s", whereClause)
	return yql
}
