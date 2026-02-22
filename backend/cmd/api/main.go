package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/yourusername/vespa-knowledge-hub/internal/models"
	"github.com/yourusername/vespa-knowledge-hub/internal/vespa"
)

type Server struct {
	vespaClient *vespa.Client
	port        string
}

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	fmt.Println("🚀 Vespa Knowledge Hub - API Server")
	fmt.Println("====================================\n")

	// Get configuration from environment variables
	vespaURL := os.Getenv("VESPA_URL")
	if vespaURL == "" {
		vespaURL = "http://localhost:8080"
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	// Initialize Vespa client
	fmt.Printf("📡 Connecting to Vespa at %s...\n", vespaURL)
	vespaClient := vespa.NewClient(vespaURL)

	// Check Vespa health
	if err := vespaClient.HealthCheck(); err != nil {
		log.Fatalf("❌ Vespa is not healthy: %v\n\nPlease start Vespa with: task vespa:start", err)
	}
	fmt.Println("✅ Connected to Vespa!\n")

	// Create server
	server := &Server{
		vespaClient: vespaClient,
		port:        port,
	}

	// Setup routes
	http.HandleFunc("/api/search", corsMiddleware(server.handleSearch))
	http.HandleFunc("/health", server.handleHealth)

	// Start server
	fmt.Printf("🌐 API Server starting on http://localhost:%s\n", port)
	fmt.Println("\nAvailable endpoints:")
	fmt.Printf("  GET  /api/search?q=<query>  - Search for code and documents\n")
	fmt.Printf("  GET  /health                - Health check\n")
	fmt.Println("\nExample:")
	fmt.Printf("  curl 'http://localhost:%s/api/search?q=function&limit=5'\n\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleSearch handles search requests
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	// Optional filters
	sourceType := r.URL.Query().Get("source_type")
	repoName := r.URL.Query().Get("repo")
	language := r.URL.Query().Get("language")

	// Pagination
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100 // Cap at 100
			}
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Create search query
	searchQuery := &models.SearchQuery{
		Query:      query,
		SourceType: sourceType,
		RepoName:   repoName,
		Language:   language,
		Limit:      limit,
		Offset:     offset,
	}

	// Execute search
	results, err := s.vespaClient.Search(searchQuery)
	if err != nil {
		log.Printf("Search error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to execute search",
		})
		return
	}

	// Log the search
	log.Printf("Search: q=%s, filters=[source_type=%s, repo=%s, language=%s], results=%d",
		query, sourceType, repoName, language, results.TotalCount)

	// Return results
	writeJSON(w, http.StatusOK, results)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Check Vespa health
	if err := s.vespaClient.HealthCheck(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
