# Automatic Semantic Versioning

This repository uses GitHub Actions to automatically create semantic version tags based on commit messages.

## How It Works

The auto-tagging workflow runs on every push to the `main` branch and:

1. **Analyzes commit messages** since the last tag to determine version increment
2. **Calculates the next version** using semantic versioning (MAJOR.MINOR.PATCH)
3. **Creates and pushes a new tag** with the calculated version
4. **Creates a GitHub Release** with changelog information

## Commit Message Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/) format for automatic version bumping:

### Version Increment Rules

| Commit Message Pattern     | Version Bump | Example                             |
| -------------------------- | ------------ | ----------------------------------- |
| `feat:` or `feature:`      | **MINOR**    | `feat: add user authentication`     |
| `fix:` or `bugfix:`        | **PATCH**    | `fix: resolve login validation bug` |
| `BREAKING CHANGE:` or `!:` | **MAJOR**    | `feat!: redesign API endpoints`     |
| Other conventional commits | **PATCH**    | `docs: update API documentation`    |
| Non-conventional commits   | **PATCH**    | `update readme file`                |

### Commit Message Examples

```bash
# Minor version bump (new features)
git commit -m "feat: add password reset functionality"
git commit -m "feature: implement user dashboard"

# Patch version bump (bug fixes)
git commit -m "fix: resolve memory leak in scanner"
git commit -m "bugfix: handle edge case in file parsing"

# Major version bump (breaking changes)
git commit -m "feat!: redesign CLI interface"
git commit -m "feat: remove deprecated API endpoints

BREAKING CHANGE: The old /api/v1 endpoints have been removed"

# Patch version bump (other changes)
git commit -m "docs: update installation guide"
git commit -m "style: fix code formatting"
git commit -m "refactor: improve error handling"
git commit -m "test: add integration tests"
git commit -m "chore: update dependencies"
```

## Workflow Features

- **Smart Detection**: Automatically detects the type of changes from commit messages
- **Skip Logic**: Won't create tags if no commits since last tag
- **Release Creation**: Automatically creates GitHub releases with changelogs
- **Conflict Prevention**: Ignores documentation-only changes (\*.md files, docs/ folder)
- **Summary Reports**: Provides detailed summaries in GitHub Actions

## Manual Tagging

If you need to create a tag manually (bypassing the automatic workflow):

```bash
# Create and push a tag manually
git tag -a v1.2.3 -m "Manual release v1.2.3"
git push origin v1.2.3
```

## Workflow Configuration

The workflow is configured to:

- **Trigger**: On push to `main` branch
- **Ignore**: Markdown files and documentation changes
- **Skip**: Commits starting with `chore(release):`
- **Permissions**: Uses `GITHUB_TOKEN` for creating tags and releases

## Version History

Tags follow semantic versioning format: `v{MAJOR}.{MINOR}.{PATCH}`

- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality additions
- **PATCH**: Backward-compatible bug fixes

## Troubleshooting

### Workflow Not Running

- Check that commits are pushed to the `main` branch
- Ensure the commit doesn't start with `chore(release):`
- Verify that changes aren't only in ignored paths (\*.md, docs/)

### Unexpected Version Increment

- Review commit messages for conventional commit format
- Check the workflow logs for detected increment type
- Multiple commit types will use the highest precedence (major > minor > patch)

### Manual Intervention

If the automatic workflow creates an incorrect tag:

1. Delete the incorrect tag: `git tag -d v1.2.3 && git push --delete origin v1.2.3`
2. Delete the GitHub release (if created)
3. Push commits with corrected messages
4. Or create the correct tag manually
