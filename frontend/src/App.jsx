import { useState } from 'react';
import { SearchBar } from './components/SearchBar';
import { ResultList } from './components/ResultList';
import './App.css';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000';

function App() {
  const [results, setResults] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [searchPerformed, setSearchPerformed] = useState(false);

  const handleSearch = async ({ query, language, repo }) => {
    // Don't search if no query and no filters
    if (!query && !language && !repo) {
      setResults(null);
      setSearchPerformed(false);
      return;
    }

    setLoading(true);
    setError(null);
    setSearchPerformed(true);

    try {
      const params = new URLSearchParams();
      if (query) params.append('q', query);
      if (language) params.append('language', language);
      if (repo) params.append('repo', repo);
      params.append('limit', '20');

      const response = await fetch(`${API_URL}/api/search?${params}`);

      if (!response.ok) {
        throw new Error(`Search failed: ${response.statusText}`);
      }

      const data = await response.json();
      setResults(data);
    } catch (err) {
      setError(err.message);
      setResults(null);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <div className="container">
          <h1 className="app-title">
            🔍 Vespa Knowledge Hub
          </h1>
          <p className="app-subtitle">
            Search your code and documents with AI-powered ranking
          </p>
        </div>
      </header>

      <main className="app-main">
        <div className="container">
          <SearchBar onSearch={handleSearch} loading={loading} />

          {searchPerformed && (
            <ResultList results={results} loading={loading} error={error} />
          )}

          {!searchPerformed && (
            <div className="welcome">
              <div className="welcome-icon">👋</div>
              <h2>Welcome to Vespa Knowledge Hub</h2>
              <p>Search across your GitHub repositories and Notion documents</p>
              <div className="features">
                <div className="feature">
                  <div className="feature-icon">⚡</div>
                  <h3>Fast Search</h3>
                  <p>Powered by Vespa search engine</p>
                </div>
                <div className="feature">
                  <div className="feature-icon">🎯</div>
                  <h3>Relevant Results</h3>
                  <p>BM25 ranking algorithm</p>
                </div>
                <div className="feature">
                  <div className="feature-icon">🔧</div>
                  <h3>Smart Filters</h3>
                  <p>Filter by language and repository</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>

      <footer className="app-footer">
        <div className="container">
          <p>
            Powered by{' '}
            <a href="https://vespa.ai" target="_blank" rel="noopener noreferrer">
              Vespa
            </a>
            {' • '}
            Built with React + Vite
          </p>
        </div>
      </footer>
    </div>
  );
}

export default App;
