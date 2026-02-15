package models

import "time"

// KnowledgeItem represents a document in the knowledge base
type KnowledgeItem struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	SourceType string    `json:"source_type"`
	SourceURL  string    `json:"source_url"`
	RepoName   string    `json:"repo_name,omitempty"`
	FilePath   string    `json:"file_path,omitempty"`
	Language   string    `json:"language,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// VespaDocument represents the Vespa document format
type VespaDocument struct {
	Fields map[string]interface{} `json:"fields"`
}

// ToVespaDocument converts KnowledgeItem to Vespa document format
func (k *KnowledgeItem) ToVespaDocument() *VespaDocument {
	return &VespaDocument{
		Fields: map[string]interface{}{
			"id":          k.ID,
			"title":       k.Title,
			"content":     k.Content,
			"source_type": k.SourceType,
			"source_url":  k.SourceURL,
			"repo_name":   k.RepoName,
			"file_path":   k.FilePath,
			"language":    k.Language,
			"created_at":  k.CreatedAt.Unix(),
			"updated_at":  k.UpdatedAt.Unix(),
		},
	}
}

// SearchResult represents a search result from Vespa
type SearchResult struct {
	TotalCount int                 `json:"total_count"`
	Hits       []SearchResultHit   `json:"hits"`
}

// SearchResultHit represents a single hit in the search results
type SearchResultHit struct {
	ID        string                 `json:"id"`
	Fields    map[string]interface{} `json:"fields"`
	Relevance float64                `json:"relevance"`
}

// SearchQuery represents a search query with filters
type SearchQuery struct {
	Query      string
	SourceType string
	RepoName   string
	Language   string
	Limit      int
	Offset     int
}
