# Releasing

This project uses [GoReleaser](https://goreleaser.com/) and GitHub Actions to automate releases.

## How to Release

1. **Tag the release:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions takes over:**
   - Builds binaries for darwin-amd64, darwin-arm64, linux-amd64
   - Creates a GitHub release with the binaries
   - Updates the Homebrew tap formula

That's it! Users can then install via any of the documented methods.

## Version Format

Use [semantic versioning](https://semver.org/):
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features, backwards compatible)
- `v1.1.1` - Patch release (bug fixes)

Pre-release versions are also supported:
- `v1.0.0-alpha.1`
- `v1.0.0-beta.1`
- `v1.0.0-rc.1`

## Setup (One-Time)

### Homebrew Tap Token

For automatic Homebrew formula updates, you need a fine-grained Personal Access Token:

1. Go to [GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens](https://github.com/settings/tokens?type=beta)

2. Click **Generate new token**

3. Configure:
   - **Token name:** `applink-homebrew-tap`
   - **Expiration:** 1 year (or your preference)
   - **Repository access:** Select "Only select repositories" → choose `jaknapp/homebrew-tap`
   - **Permissions:** 
     - Contents: Read and write

4. Generate and copy the token

5. Add to this repo's secrets:
   - Go to `jaknapp/applink` → Settings → Secrets and variables → Actions
   - Add secret: `HOMEBREW_TAP_GITHUB_TOKEN` with the token value

### Why a Fine-Grained PAT?

The built-in `GITHUB_TOKEN` only has access to the current repository. To update the `homebrew-tap` repo, we need cross-repo access. A fine-grained PAT is the recommended approach because:

- It's scoped to only the `homebrew-tap` repo (not all your repos)
- It only has the permissions needed (contents write)
- It can be set to expire

## Local Testing

Test the release process locally without publishing:

```bash
# Install goreleaser
brew install goreleaser

# Test build (creates binaries in dist/)
goreleaser build --snapshot --clean

# Test full release (dry run)
goreleaser release --snapshot --clean
```

## What Gets Built

| Platform | Architecture | Binary |
|----------|--------------|--------|
| macOS    | Intel (amd64) | `applink_darwin_amd64.tar.gz` |
| macOS    | Apple Silicon (arm64) | `applink_darwin_arm64.tar.gz` |
| Linux    | x86_64 (amd64) | `applink_linux_amd64.tar.gz` |

## Troubleshooting

### Homebrew formula not updating

1. Check the Actions log for errors
2. Verify `HOMEBREW_TAP_GITHUB_TOKEN` is set and not expired
3. Ensure the token has write access to `jaknapp/homebrew-tap`

### Build fails

```bash
# Check goreleaser config is valid
goreleaser check
```
