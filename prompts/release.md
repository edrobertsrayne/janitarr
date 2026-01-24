# Prompt: Create GitHub Release

Create a new GitHub release for this project using semantic versioning. Follow these steps:

## 1. Analyze Git History

- Get the most recent git tag: `git describe --tags --abbrev=0 2>/dev/null`
- If no tags exist, analyze all commits; otherwise analyze commits since the last tag
- Parse conventional commit prefixes to categorize changes

## 2. Determine Version Bump

Based on conventional commits since the last tag:

- **MAJOR** (breaking): Commits containing `BREAKING CHANGE:` in body or `!` after type (e.g., `feat!:`)
- **MINOR**: Any `feat:` or `feat(scope):` commits
- **PATCH**: Any `fix:`, `perf:`, or other non-feature changes

If no previous tag exists, start at v0.1.0.

## 3. Generate Release Notes

Group commits by type using these categories:

- 🚀 **Features** - `feat:` commits
- 🐛 **Bug Fixes** - `fix:` commits
- ⚡ **Performance** - `perf:` commits
- 📚 **Documentation** - `docs:` commits
- 🔧 **Build/CI** - `build:`, `ci:` commits
- ♻️ **Refactoring** - `refactor:` commits
- 🧪 **Tests** - `test:` commits
- 🎨 **Style** - `style:` commits
- 🗑️ **Chores** - `chore:` commits

Format each entry as: `- <description> (<short-hash>)`

Remove the type prefix from descriptions for cleaner notes.

## 4. Create the Release

Use GitHub CLI to create the tag and release together:

```bash
gh release create vX.Y.Z --title "vX.Y.Z" --notes "<generated notes>"
```

This will create both the git tag and GitHub release in one command.

## 5. Verification

After creating the release:

- Run `git describe --tags` to confirm the tag is set
- Confirm the release is visible at the GitHub releases page

## Important Notes

- This project injects version via build-time ldflags from `git describe --tags`, so no source code updates are needed
- Do NOT push force or modify existing tags without explicit permission
- Ask me before creating the release to confirm the version number and notes look correct
