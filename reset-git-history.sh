#!/bin/bash

# Script to reset Git history and remove personal information
# Usage: ./reset-git-history.sh

set -e

cd "$(dirname "$0")"

echo "🔄 Resetting Git history..."

# Remove existing Git history
if [ -d ".git" ]; then
    echo "📦 Removing existing Git history..."
    rm -rf .git
fi

# Initialize fresh repository
echo "🆕 Initializing new Git repository..."
git init
git branch -M main

# Configure Git user for this repository
echo "⚙️  Configuring Git user..."
git config user.name "KeyClaim"
git config user.email "support@keyclaim.org"

# Add all files
echo "📝 Adding files..."
git add .

# Create initial commit
echo "💾 Creating initial commit..."
git commit -m "Initial commit: KeyClaim Go SDK v1.0.0"

# Add remote
echo "🔗 Setting up remote..."
if git remote get-url origin &>/dev/null; then
    git remote set-url origin https://github.com/creasoftlb/keyclaim-go-sdk.git
else
    git remote add origin https://github.com/creasoftlb/keyclaim-go-sdk.git
fi

# Show what will be pushed
echo ""
echo "📊 Commit to be pushed:"
git log --oneline --format="%h %an <%ae> %s"

echo ""
echo "⚠️  WARNING: This will overwrite remote history!"
echo "   Anyone who has cloned will need to re-clone."
read -p "Continue with force push? (y/N) " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🚀 Pushing to remote..."
    git push -u origin main --force
    echo ""
    echo "✅ History reset complete!"
    echo ""
    echo "📌 Next steps:"
    echo "   1. Create a release tag: git tag v1.0.0 && git push origin v1.0.0"
    echo "   2. Create GitHub release at: https://github.com/creasoftlb/keyclaim-go-sdk/releases/new"
else
    echo "❌ Aborted. Local repository is ready, but not pushed."
    echo "   Run 'git push -u origin main --force' when ready."
fi

