package github

import "time"

// RepositoryInfo represents basic information about a GitHub repository
type RepositoryInfo struct {
	Owner         string
	Name          string
	FullName      string
	Description   string
	DefaultBranch string
	Language      string
	StarCount     int
	ForkCount     int
	UpdatedAt     time.Time
}

// FileMetadata represents metadata about a file in a repository
type FileMetadata struct {
	Path      string
	SHA       string
	Size      int
	Language  string
	UpdatedAt time.Time
}

// IndexStats represents statistics about the indexing process
type IndexStats struct {
	TotalFiles     int
	IndexedFiles   int
	SkippedFiles   int
	FailedFiles    int
	TotalBytes     int64
	StartTime      time.Time
	EndTime        time.Time
	RepositoryName string
}

// NewIndexStats creates a new IndexStats instance
func NewIndexStats(repoName string) *IndexStats {
	return &IndexStats{
		RepositoryName: repoName,
		StartTime:      time.Now(),
	}
}

// AddIndexed increments the indexed files counter
func (s *IndexStats) AddIndexed(size int) {
	s.IndexedFiles++
	s.TotalBytes += int64(size)
}

// AddSkipped increments the skipped files counter
func (s *IndexStats) AddSkipped() {
	s.SkippedFiles++
}

// AddFailed increments the failed files counter
func (s *IndexStats) AddFailed() {
	s.FailedFiles++
}

// Finish marks the indexing as complete
func (s *IndexStats) Finish() {
	s.EndTime = time.Now()
}

// Duration returns the duration of the indexing process
func (s *IndexStats) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}
