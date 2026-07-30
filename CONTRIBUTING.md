# Contributing

## Commit convention — required for automated releases

Releases (versioning, `CHANGELOG.md`, GitHub Releases) are fully automated by
[release-please](https://github.com/googleapis/release-please) based on
commit messages. Every commit to `main` must follow
[Conventional Commits](https://www.conventionalcommits.org/):

```
feat(infra-maps-api): add Prometheus scraper
fix(infra-maps-ui): correct status badge color for warning state
feat(infra-maps-api)!: change /tree response shape (BREAKING CHANGE)
docs: update ROADMAP
chore: bump dependency
```

| Prefix | Effect on version |
|--------|--------------------|
| `fix:` | patch bump |
| `feat:` | minor bump |
| `feat!:` / `BREAKING CHANGE:` in footer | major bump |
| `docs:`, `chore:`, `test:`, `ci:`, `refactor:` | no release, included in changelog under their own section |

Scope the commit to the package it touches (`infra-maps-api` or
`infra-maps-ui`) so release-please versions each package independently — this
repo is a monorepo with two release manifests (see
`release-please-config.json`).

## Before opening a PR

- `cd infra-maps-api && go vet ./... && go test ./... -race`
- `cd infra-maps-ui && pnpm lint && pnpm test && pnpm build`

Both run automatically in CI (`.github/workflows/backend-ci.yml`,
`.github/workflows/frontend-ci.yml`) and must pass before merge.
