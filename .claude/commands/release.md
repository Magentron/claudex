# Release Command

You are performing an npm release for claudex. Follow this process step by step.

## Step 1: Analyze Current State

Run these commands to understand the current release state:

```bash
# Get the latest release tag
git describe --tags --abbrev=0 2>/dev/null || echo "No tags found"

# Get current version in npm/version.txt
cat npm/version.txt

# Show commits since last release (if tags exist)
git log $(git describe --tags --abbrev=0 2>/dev/null)..HEAD --oneline 2>/dev/null || git log --oneline -20
```

## Step 2: Determine Version Bump

Based on the commits since the last release, determine the appropriate version bump:

- **PATCH** (0.0.X): Bug fixes, documentation, minor improvements
- **MINOR** (0.X.0): New features, backward-compatible changes
- **MAJOR** (X.0.0): Breaking changes, major rewrites

Look for commit prefixes:
- `fix:` → patch
- `feat:` → minor
- `BREAKING CHANGE` or `!:` → major

## Step 3: Ask User for Confirmation

Before proceeding, ask the user:
1. Show them the commits that will be included
2. Suggest the version bump type (patch/minor/major)
3. Show the proposed new version number
4. Get explicit approval to proceed

## Step 4: Execute Release

Once approved:

```bash
# 1. Update version.txt with new version
echo "NEW_VERSION" > npm/version.txt

# 2. Commit the version bump
git add npm/version.txt
git commit -m "chore: release vNEW_VERSION"

# 3. Push the commit
git push origin main

# 4. Create and push the tag (this triggers GitHub Actions)
git tag vNEW_VERSION
git push origin vNEW_VERSION
```

## Step 5: Monitor Release

After pushing the tag:
1. Provide the GitHub Actions URL: https://github.com/mgonzalezbaile/claudex/actions
2. Tell the user to monitor the workflow
3. Once complete, verify with: `npm view @claudex/cli`

## Important Notes

- The GitHub Actions workflow builds binaries for all platforms and publishes to npm
- All 5 packages are published: @claudex/cli, @claudex/darwin-arm64, @claudex/darwin-x64, @claudex/linux-x64, @claudex/linux-arm64
- If a release fails, you may need to bump the version again (can't republish same version)
- Always ensure the main branch is clean before releasing

## Troubleshooting

If the release fails:
1. Check GitHub Actions logs for errors
2. Common issues:
   - NPM_TOKEN expired → user needs to regenerate and update GitHub secret
   - Version already exists → bump to next version
   - Build failure → check Go compilation errors
