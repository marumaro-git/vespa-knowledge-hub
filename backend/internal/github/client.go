package github

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client represents a GitHub API client
type Client struct {
	client *github.Client
	ctx    context.Context
}

// NewClient creates a new GitHub client with authentication
func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

// Repository represents a GitHub repository
type Repository struct {
	Owner string
	Name  string
}

// ParseRepository parses "owner/repo" format into Repository
func ParseRepository(repo string) (*Repository, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}
	return &Repository{
		Owner: parts[0],
		Name:  parts[1],
	}, nil
}

// FileItem represents a file in a repository
type FileItem struct {
	Path     string
	Content  string
	SHA      string
	Language string
}

// GetRepositoryFiles retrieves all files from a repository
func (c *Client) GetRepositoryFiles(repo *Repository) ([]*FileItem, error) {
	var files []*FileItem

	// Get the default branch
	ghRepo, _, err := c.client.Repositories.Get(c.ctx, repo.Owner, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	defaultBranch := ghRepo.GetDefaultBranch()
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	// Get the tree recursively
	tree, _, err := c.client.Git.GetTree(c.ctx, repo.Owner, repo.Name, defaultBranch, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	// Filter for files only (not directories)
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			path := entry.GetPath()

			// Skip binary files and large files
			if c.shouldSkipFile(path) {
				continue
			}

			// Get file content
			content, err := c.GetFileContent(repo, path, defaultBranch)
			if err != nil {
				// Log error but continue with other files
				fmt.Printf("Warning: failed to get content for %s: %v\n", path, err)
				continue
			}

			files = append(files, &FileItem{
				Path:     path,
				Content:  content,
				SHA:      entry.GetSHA(),
				Language: c.detectLanguage(path),
			})
		}
	}

	return files, nil
}

// GetFileContent retrieves the content of a specific file
func (c *Client) GetFileContent(repo *Repository, path, ref string) (string, error) {
	fileContent, _, _, err := c.client.Repositories.GetContents(
		c.ctx,
		repo.Owner,
		repo.Name,
		path,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get file content: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode content: %w", err)
	}

	return content, nil
}

// shouldSkipFile determines if a file should be skipped during indexing
func (c *Client) shouldSkipFile(path string) bool {
	// Skip common binary and non-text files
	skipExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".ico", ".svg",
		".pdf", ".zip", ".tar", ".gz", ".7z", ".rar",
		".exe", ".dll", ".so", ".dylib",
		".mp3", ".mp4", ".avi", ".mov",
		".woff", ".woff2", ".ttf", ".eot",
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, skipExt := range skipExtensions {
		if ext == skipExt {
			return true
		}
	}

	// Skip common directories
	skipPatterns := []string{
		"node_modules/",
		"vendor/",
		".git/",
		"dist/",
		"build/",
		"target/",
		".next/",
		"__pycache__/",
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

// detectLanguage detects programming language from file extension
func (c *Client) detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	languageMap := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".jsx":   "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".py":    "python",
		".java":  "java",
		".c":     "c",
		".cpp":   "cpp",
		".cc":    "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".rs":    "rust",
		".rb":    "ruby",
		".php":   "php",
		".swift": "swift",
		".kt":    "kotlin",
		".scala": "scala",
		".sh":    "shell",
		".bash":  "shell",
		".zsh":   "shell",
		".fish":  "shell",
		".sql":   "sql",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".sass":  "sass",
		".md":    "markdown",
		".json":  "json",
		".xml":   "xml",
		".yaml":  "yaml",
		".yml":   "yaml",
		".toml":  "toml",
		".ini":   "ini",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	return "text"
}

// GetRateLimit returns the current rate limit status
func (c *Client) GetRateLimit() (*github.RateLimits, error) {
	limits, _, err := c.client.RateLimits(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limits: %w", err)
	}
	return limits, nil
}
