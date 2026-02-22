package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourusername/vespa-knowledge-hub/internal/github"
	"github.com/yourusername/vespa-knowledge-hub/internal/models"
	"github.com/yourusername/vespa-knowledge-hub/internal/vespa"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	fmt.Println("🚀 Vespa Knowledge Hub - GitHub Indexer")
	fmt.Println("========================================\n")

	// Get configuration from environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("❌ GITHUB_TOKEN is not set. Please set it with: export GITHUB_TOKEN=ghp_your_token")
	}

	targetRepos := os.Getenv("TARGET_REPOS")
	if targetRepos == "" {
		log.Fatal("❌ TARGET_REPOS is not set. Please set it with: export TARGET_REPOS=owner/repo1,owner/repo2")
	}

	vespaURL := os.Getenv("VESPA_URL")
	if vespaURL == "" {
		vespaURL = "http://localhost:8080"
	}

	// Initialize clients
	fmt.Println("📡 Initializing clients...")
	ghClient := github.NewClient(githubToken)
	vespaClient := vespa.NewClient(vespaURL)

	// Check Vespa health
	fmt.Printf("🔍 Checking Vespa health at %s...\n", vespaURL)
	if err := vespaClient.HealthCheck(); err != nil {
		log.Fatalf("❌ Vespa is not healthy: %v\n\nPlease start Vespa with: task vespa:start", err)
	}
	fmt.Println("✅ Vespa is healthy!\n")

	// Check GitHub rate limit
	rateLimit, err := ghClient.GetRateLimit()
	if err != nil {
		log.Printf("⚠️  Warning: Could not check GitHub rate limit: %v", err)
	} else {
		fmt.Printf("📊 GitHub API Rate Limit: %d/%d remaining\n\n",
			rateLimit.Core.Remaining, rateLimit.Core.Limit)
	}

	// Parse repositories
	repos := strings.Split(targetRepos, ",")
	fmt.Printf("📚 Indexing %d repositories:\n", len(repos))
	for i, repo := range repos {
		fmt.Printf("  %d. %s\n", i+1, strings.TrimSpace(repo))
	}
	fmt.Println()

	// Index each repository
	totalStats := struct {
		TotalFiles   int
		IndexedFiles int
		SkippedFiles int
		FailedFiles  int
		TotalBytes   int64
	}{}

	for _, repoStr := range repos {
		repoStr = strings.TrimSpace(repoStr)
		fmt.Printf("\n📦 Processing repository: %s\n", repoStr)
		fmt.Println(strings.Repeat("=", 60))

		// Parse repository
		repo, err := github.ParseRepository(repoStr)
		if err != nil {
			log.Printf("❌ Failed to parse repository %s: %v", repoStr, err)
			continue
		}

		// Index repository
		stats := indexRepository(ghClient, vespaClient, repo)

		// Update total stats
		totalStats.TotalFiles += stats.TotalFiles
		totalStats.IndexedFiles += stats.IndexedFiles
		totalStats.SkippedFiles += stats.SkippedFiles
		totalStats.FailedFiles += stats.FailedFiles
		totalStats.TotalBytes += stats.TotalBytes

		// Print repository stats
		fmt.Printf("\n📊 Repository Stats:\n")
		fmt.Printf("  ✅ Indexed: %d files\n", stats.IndexedFiles)
		fmt.Printf("  ⏭️  Skipped: %d files\n", stats.SkippedFiles)
		if stats.FailedFiles > 0 {
			fmt.Printf("  ❌ Failed: %d files\n", stats.FailedFiles)
		}
		fmt.Printf("  📏 Total size: %.2f MB\n", float64(stats.TotalBytes)/(1024*1024))
		fmt.Printf("  ⏱️  Duration: %s\n", stats.Duration().Round(time.Second))
	}

	// Print overall summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 Indexing Complete!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📊 Overall Stats:\n")
	fmt.Printf("  📚 Repositories: %d\n", len(repos))
	fmt.Printf("  ✅ Indexed: %d files\n", totalStats.IndexedFiles)
	fmt.Printf("  ⏭️  Skipped: %d files\n", totalStats.SkippedFiles)
	if totalStats.FailedFiles > 0 {
		fmt.Printf("  ❌ Failed: %d files\n", totalStats.FailedFiles)
	}
	fmt.Printf("  📏 Total size: %.2f MB\n", float64(totalStats.TotalBytes)/(1024*1024))
	fmt.Println()
	fmt.Println("✨ You can now search your code!")
	fmt.Println("   Start the API server with: task backend:run")
	fmt.Println("   Or search directly:")
	fmt.Printf("   curl 'http://localhost:8080/search/?yql=select * from knowledge_item where userQuery()&query=your_search'\n")
}

// indexRepository indexes all files from a repository
func indexRepository(ghClient *github.Client, vespaClient *vespa.Client, repo *github.Repository) *github.IndexStats {
	stats := github.NewIndexStats(fmt.Sprintf("%s/%s", repo.Owner, repo.Name))

	fmt.Printf("📥 Fetching files from GitHub...\n")
	files, err := ghClient.GetRepositoryFiles(repo)
	if err != nil {
		log.Printf("❌ Failed to get repository files: %v", err)
		stats.Finish()
		return stats
	}

	stats.TotalFiles = len(files)
	fmt.Printf("📄 Found %d files\n", stats.TotalFiles)

	if stats.TotalFiles == 0 {
		fmt.Println("⚠️  No files found in repository")
		stats.Finish()
		return stats
	}

	fmt.Println("📤 Indexing files to Vespa...")

	// Progress counter
	progressInterval := 10
	if stats.TotalFiles > 100 {
		progressInterval = stats.TotalFiles / 10
	}

	for i, file := range files {
		// Create knowledge item
		docID := fmt.Sprintf("gh_%s_%s_%s",
			repo.Owner,
			repo.Name,
			strings.ReplaceAll(file.Path, "/", "_"),
		)

		item := &models.KnowledgeItem{
			ID:         docID,
			Title:      file.Path,
			Content:    file.Content,
			SourceType: "github_code",
			SourceURL: fmt.Sprintf("https://github.com/%s/%s/blob/main/%s",
				repo.Owner, repo.Name, file.Path),
			RepoName:  fmt.Sprintf("%s/%s", repo.Owner, repo.Name),
			FilePath:  file.Path,
			Language:  file.Language,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Index to Vespa
		if err := vespaClient.IndexDocument(item); err != nil {
			log.Printf("  ❌ Failed to index %s: %v", file.Path, err)
			stats.AddFailed()
			continue
		}

		stats.AddIndexed(len(file.Content))

		// Show progress
		if (i+1)%progressInterval == 0 || i == len(files)-1 {
			fmt.Printf("  Progress: %d/%d (%.1f%%)\n",
				i+1, stats.TotalFiles,
				float64(i+1)/float64(stats.TotalFiles)*100)
		}
	}

	stats.Finish()
	return stats
}
