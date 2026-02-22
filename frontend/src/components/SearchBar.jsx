import { useState } from 'react';
import './SearchBar.css';

export function SearchBar({ onSearch, loading }) {
  const [query, setQuery] = useState('');
  const [language, setLanguage] = useState('');
  const [repo, setRepo] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    onSearch({ query, language, repo });
  };

  const handleClear = () => {
    setQuery('');
    setLanguage('');
    setRepo('');
    onSearch({ query: '', language: '', repo: '' });
  };

  return (
    <div className="search-bar">
      <form onSubmit={handleSubmit}>
        <div className="search-input-group">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search code and documents..."
            className="search-input"
            autoFocus
          />
          <button type="submit" className="search-button" disabled={loading}>
            {loading ? '🔄' : '🔍'} Search
          </button>
        </div>

        <div className="filters">
          <div className="filter-group">
            <label htmlFor="language">Language:</label>
            <select
              id="language"
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
              className="filter-select"
            >
              <option value="">All</option>
              <option value="go">Go</option>
              <option value="javascript">JavaScript</option>
              <option value="typescript">TypeScript</option>
              <option value="python">Python</option>
              <option value="java">Java</option>
              <option value="rust">Rust</option>
              <option value="cpp">C++</option>
              <option value="markdown">Markdown</option>
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="repo">Repository:</label>
            <input
              id="repo"
              type="text"
              value={repo}
              onChange={(e) => setRepo(e.target.value)}
              placeholder="owner/repo"
              className="filter-input"
            />
          </div>

          {(query || language || repo) && (
            <button type="button" onClick={handleClear} className="clear-button">
              Clear
            </button>
          )}
        </div>
      </form>
    </div>
  );
}
