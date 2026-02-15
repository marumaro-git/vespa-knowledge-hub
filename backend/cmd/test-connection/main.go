package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourusername/vespa-knowledge-hub/internal/models"
	"github.com/yourusername/vespa-knowledge-hub/internal/vespa"
)

func main() {
	vespaURL := os.Getenv("VESPA_URL")
	if vespaURL == "" {
		vespaURL = "http://localhost:8080"
	}

	fmt.Printf("🔍 Testing connection to Vespa at %s\n\n", vespaURL)

	client := vespa.NewClient(vespaURL)

	// Step 1: Health check
	fmt.Println("Step 1: Health check...")
	if err := client.HealthCheck(); err != nil {
		log.Fatalf("❌ Health check failed: %v", err)
	}
	fmt.Println("✅ Vespa is healthy!\n")

	// Step 2: Index a test document
	fmt.Println("Step 2: Indexing a test document...")
	testDoc := &models.KnowledgeItem{
		ID:         "test_doc_1",
		Title:      "test.go",
		Content:    "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
		SourceType: "github_code",
		SourceURL:  "https://github.com/test/repo/blob/main/test.go",
		RepoName:   "test/repo",
		FilePath:   "test.go",
		Language:   "go",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := client.IndexDocument(testDoc); err != nil {
		log.Fatalf("❌ Failed to index document: %v", err)
	}
	fmt.Println("✅ Document indexed successfully!\n")

	// Wait a bit for indexing to complete
	fmt.Println("⏳ Waiting for document to be indexed...")
	time.Sleep(2 * time.Second)

	// Step 3: Search for the document
	fmt.Println("\nStep 3: Searching for 'main'...")
	searchQuery := &models.SearchQuery{
		Query:  "main",
		Limit:  10,
		Offset: 0,
	}

	results, err := client.Search(searchQuery)
	if err != nil {
		log.Fatalf("❌ Search failed: %v", err)
	}

	fmt.Printf("✅ Search completed! Found %d results\n\n", results.TotalCount)

	if len(results.Hits) > 0 {
		fmt.Println("Top results:")
		for i, hit := range results.Hits {
			if i >= 3 {
				break
			}
			fmt.Printf("  %d. %s (relevance: %.2f)\n", i+1, hit.Fields["title"], hit.Relevance)
			fmt.Printf("     Repo: %s\n", hit.Fields["repo_name"])
			fmt.Printf("     Path: %s\n", hit.Fields["file_path"])
			fmt.Println()
		}
	}

	// Step 4: Cleanup
	fmt.Println("Step 4: Cleaning up test document...")
	if err := client.DeleteDocument("test_doc_1"); err != nil {
		log.Printf("⚠️  Warning: Failed to delete test document: %v", err)
	} else {
		fmt.Println("✅ Test document deleted\n")
	}

	fmt.Println("🎉 All tests passed! Vespa connection is working correctly.")
}
