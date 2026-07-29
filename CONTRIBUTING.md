# Contributing to VortexUiPro

First off, thank you for considering contributing! 🎉

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Development Process

### 1. Fork & Clone

```bash
git clone https://github.com/your-username/VortexUiPro.git
cd VortexUiPro
git remote add upstream https://github.com/iPmartNetwork/VortexUiPro.git
```

### 2. Setup Environment

```bash
# Backend
go mod download
go build ./...

# Frontend
cd web
npm ci
npm run build -- --mode production
cd ..
```

### 3. Create a Branch

```bash
git checkout -b feat/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming:
- `feat/*` — New features
- `fix/*` — Bug fixes
- `docs/*` — Documentation
- `refactor/*` — Code refactoring
- `test/*` — Adding tests
- `perf/*` — Performance improvements

### 4. Make Changes

**Code Style:**
- Go: Follow `gofmt` / `go vet` standards
- TypeScript: Follow the project's ESLint config
- Run `gofmt -s -w ./internal/` before committing

**Commit Messages:**
```
feat: add structured logger with zerolog
fix: resolve port conflict detection edge case
docs: update API reference
test: add inbound service unit tests
```

We follow [Conventional Commits](https://www.conventionalcommits.org/).

### 5. Test

```bash
# Run all Go tests
go test ./internal/...

# Run frontend type check
cd web && npx tsc --noEmit
```

### 6. Push & PR

```bash
git push origin feat/your-feature-name
```

Then open a Pull Request against the `master` branch.

## Pull Request Guidelines

- **One PR = one feature/fix.** Keep changes focused.
- **Write tests** for new functionality.
- **Update documentation** (README, API docs) if needed.
- **Keep the PR description clear** — what, why, and how.
- **Reference related issues** in the description.

## Project Structure

```
internal/
├── service/     # Business logic (41 services)
├── api/         # HTTP handlers & middleware
├── core/        # Engine drivers (Xray, Sing-box)
├── cluster/     # Multi-node mesh
├── database/    # GORM models & repository
├── events/      # Event bus
├── metrics/     # Prometheus collector
├── monitor/     # Health check engine
├── rbac/        # Role-based access
├── config/      # Environment config
└── domain/      # Domain types

web/src/
├── pages/       # 40 React pages
├── components/  # Reusable components
└── locales/     # 10 languages
```

## Getting Help

- Open a [Discussion](https://github.com/iPmartNetwork/VortexUiPro/discussions)
- Join our [Telegram Group](https://t.me/VortexUiPro)

Thank you for contributing! 🚀
