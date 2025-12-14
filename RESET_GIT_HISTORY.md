# Reset Git History and Remove Personal Information

This guide will help you reset the Git history and remove your name from commits.

## Option 1: Complete Reset (Recommended for New Repository)

This will completely remove all commit history and start fresh:

```bash
cd /Users/shogun/Desktop/SDKS/keyclaim-go-sdk

# Remove existing Git history
rm -rf .git

# Initialize new Git repository
git init

# Add all files
git add .

# Create initial commit with generic author
git commit -m "Initial commit" --author="KeyClaim <support@keyclaim.org>"

# Add remote (if not already added)
git remote add origin https://github.com/creasoftlb/keyclaim-go-sdk.git

# Force push to replace history (WARNING: This will overwrite remote history)
git push -u origin main --force
```

## Option 2: Rewrite History (Keep Some Commits)

If you want to rewrite existing commits to change author information:

```bash
cd /Users/shogun/Desktop/SDKS/keyclaim-go-sdk

# Set the new author for all commits
git filter-branch --env-filter '
export GIT_AUTHOR_NAME="KeyClaim"
export GIT_AUTHOR_EMAIL="support@keyclaim.org"
export GIT_COMMITTER_NAME="KeyClaim"
export GIT_COMMITTER_EMAIL="support@keyclaim.org"
' --tag-name-filter cat -- --branches --tags

# Force push to replace history
git push origin main --force
```

## Option 3: Using git-filter-repo (More Modern Approach)

If you have `git-filter-repo` installed:

```bash
cd /Users/shogun/Desktop/SDKS/keyclaim-go-sdk

# Install git-filter-repo if needed
# pip install git-filter-repo

# Rewrite all commits with new author
git filter-repo --name "KeyClaim" --email "support@keyclaim.org" --force

# Force push
git push origin main --force
```

## Option 4: Start Fresh Branch

Create a new orphan branch (no history):

```bash
cd /Users/shogun/Desktop/SDKS/keyclaim-go-sdk

# Create new orphan branch
git checkout --orphan new-main

# Remove all files from staging
git rm -rf .

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit" --author="KeyClaim <support@keyclaim.org>"

# Delete old main branch
git branch -D main

# Rename current branch to main
git branch -m main

# Force push
git push origin main --force
```

## Recommended: Complete Reset Script

Here's a complete script to reset everything:

```bash
#!/bin/bash
cd /Users/shogun/Desktop/SDKS/keyclaim-go-sdk

# Remove Git history
rm -rf .git

# Initialize fresh repository
git init
git branch -M main

# Configure Git user for this repository
git config user.name "KeyClaim"
git config user.email "support@keyclaim.org"

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: KeyClaim Go SDK v1.0.0"

# Add remote
git remote add origin https://github.com/creasoftlb/keyclaim-go-sdk.git 2>/dev/null || git remote set-url origin https://github.com/creasoftlb/keyclaim-go-sdk.git

# Force push (WARNING: This overwrites remote history)
echo "⚠️  WARNING: This will overwrite remote history!"
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git push -u origin main --force
    echo "✅ History reset complete!"
else
    echo "❌ Aborted"
fi
```

## Important Warnings

⚠️ **Force Push Warning:**
- `--force` will overwrite the remote repository history
- Anyone who has cloned the repository will need to re-clone
- This cannot be undone easily
- Make sure you have backups if needed

⚠️ **Before Force Pushing:**
- Make sure you have all your code committed locally
- Consider backing up the repository
- Inform collaborators if any
- Check that you have the correct remote URL

## After Resetting

1. Create a release tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. Create GitHub release:
   - Go to repository → Releases → Create new release
   - Tag: v1.0.0
   - Add release notes

## Verify

After pushing, verify the commit author:

```bash
git log --format="%an <%ae>"
```

Should show: `KeyClaim <support@keyclaim.org>`

