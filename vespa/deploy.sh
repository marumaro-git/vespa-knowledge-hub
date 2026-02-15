#!/bin/bash

set -e

echo "🚀 Deploying Vespa application..."

# Check if vespa-cli is installed
if ! command -v vespa &> /dev/null; then
    echo "❌ vespa-cli is not installed"
    echo "Install with: brew install vespa-cli"
    exit 1
fi

# Set target to local
vespa config set target local

# Deploy the application
echo "📦 Deploying application package..."
vespa deploy --wait 300

echo "✅ Deployment complete!"
echo ""
echo "🔍 Health check:"
curl -s http://localhost:8080/state/v1/health | jq .

echo ""
echo "📊 Application status:"
vespa status
