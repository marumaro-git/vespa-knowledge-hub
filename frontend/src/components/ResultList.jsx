import './ResultList.css';

export function ResultList({ results, loading, error }) {
  if (loading) {
    return (
      <div className="result-list-empty">
        <div className="loading-spinner">🔄</div>
        <p>Searching...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="result-list-empty error">
        <div className="error-icon">⚠️</div>
        <h3>Search Error</h3>
        <p>{error}</p>
      </div>
    );
  }

  if (!results || results.total_count === 0) {
    return (
      <div className="result-list-empty">
        <div className="empty-icon">🔍</div>
        <h3>No results found</h3>
        <p>Try adjusting your search query or filters</p>
      </div>
    );
  }

  return (
    <div className="result-list">
      <div className="result-count">
        Found <strong>{results.total_count}</strong> result{results.total_count !== 1 ? 's' : ''}
      </div>

      <div className="results">
        {results.hits.map((hit, index) => (
          <ResultItem key={hit.id || index} hit={hit} />
        ))}
      </div>
    </div>
  );
}

function ResultItem({ hit }) {
  const { fields, relevance } = hit;
  const {
    title,
    content,
    source_url,
    repo_name,
    file_path,
    language,
    source_type
  } = fields;

  // Truncate content for preview
  const contentPreview = content?.length > 300
    ? content.substring(0, 300) + '...'
    : content;

  // Get language badge color
  const getLanguageColor = (lang) => {
    const colors = {
      go: '#00ADD8',
      javascript: '#F7DF1E',
      typescript: '#3178C6',
      python: '#3776AB',
      java: '#007396',
      rust: '#CE422B',
      cpp: '#00599C',
      markdown: '#083FA1',
    };
    return colors[lang] || '#6B7280';
  };

  return (
    <div className="result-item">
      <div className="result-header">
        <h3 className="result-title">
          <a href={source_url} target="_blank" rel="noopener noreferrer">
            {title || file_path}
          </a>
        </h3>
        <div className="result-meta">
          {language && (
            <span
              className="language-badge"
              style={{ backgroundColor: getLanguageColor(language) }}
            >
              {language}
            </span>
          )}
          <span className="relevance-score" title="Relevance score">
            {(relevance * 100).toFixed(0)}%
          </span>
        </div>
      </div>

      <div className="result-info">
        {repo_name && (
          <span className="repo-name">
            📁 {repo_name}
          </span>
        )}
        {file_path && repo_name && <span className="separator">•</span>}
        {file_path && (
          <span className="file-path">
            📄 {file_path}
          </span>
        )}
      </div>

      {content && (
        <pre className="code-preview">
          <code>{contentPreview}</code>
        </pre>
      )}

      {source_url && (
        <a
          href={source_url}
          target="_blank"
          rel="noopener noreferrer"
          className="view-source"
        >
          View Source →
        </a>
      )}
    </div>
  );
}
