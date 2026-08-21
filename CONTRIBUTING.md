# Contributing to `terraform-provider-seerr`

Thank you for your interest in contributing! We welcome contributions of all kinds: new resources, bug fixes, documentation improvements, test enhancements, and feedback.

This project supports both **OpenTofu** and **HashiCorp Terraform** as first-class Infrastructure as Code engines.

---

## 🧭 Code of Conduct

All contributors and maintainers are expected to adhere to our [Code of Conduct](CODE_OF_CONDUCT.md). Please be respectful and constructive in all interactions.

---

## 🛠️ Prerequisites & Setup

### Requirements
- **Go**: Version `1.24` or newer
- **OpenTofu** (`>= 1.8.0`) or **HashiCorp Terraform** (`>= 1.5.0`)
- **Docker** (optional, for running local test container environments)
- **Git**

### 1-Click Dev Container
If you use VS Code or GitHub Codespaces, you can open this repository directly in the provided Dev Container (`.devcontainer/devcontainer.json`), which comes with Go, OpenTofu, Terraform, and all tooling pre-configured.

### Local Clone & Build
```bash
# Clone the repository
git clone https://github.com/Josh-Archer/terraform-provider-seerr.git
cd terraform-provider-seerr

# Download Go modules
go mod download

# Verify compilation
go build ./...
```

---

## 🚀 Step-by-Step: Your First PR

1. **Find an Issue**:
   - Check open issues labeled [`good-first-issue`](https://github.com/Josh-Archer/terraform-provider-seerr/labels/good-first-issue) or [`enhancement`](https://github.com/Josh-Archer/terraform-provider-seerr/labels/enhancement).
   - If proposing a new feature or resource, please open a [Feature Request Issue](https://github.com/Josh-Archer/terraform-provider-seerr/issues/new/choose) first to discuss the design.

2. **Create a Branch**:
   ```bash
   git checkout -b feat/my-new-feature
   # or
   git checkout -b fix/issue-123
   ```

3. **Make Your Changes**:
   - Resource implementations reside in `internal/provider/resource_*.go`.
   - Data source implementations reside in `internal/provider/data_source_*.go`.
   - All resources and data sources must be registered in `internal/provider/registry.go`.
   - Documentation templates reside in `templates/resources/*.md.tmpl` and `templates/data-sources/*.md.tmpl`.

4. **Run Local Validation (Fast Gate)**:
   ```bash
   # Format code
   gofmt -s -w .

   # Regenerate documentation and schema bindings
   go generate ./...

   # Run local validation gate
   bash ./scripts/test-all-locally.sh
   ```

5. **Commit Your Changes**:
   We use [Conventional Commits](https://www.conventionalcommits.org/) to automatically generate release notes and trigger automated versioning.
   ```bash
   git commit -m "feat: add seerr_custom_resource (#123)"
   # or
   git commit -m "fix: handle nil response in user quota read (#124)"
   ```

6. **Submit a Pull Request**:
   - Push your branch and open a Pull Request against `main`.
   - Complete the checklist in our PR template. CI will automatically run unit tests, OpenTofu integration tests, OpenAPI coverage checks, and linters.

---

## 🧪 Testing Guidelines

### Unit Tests
Every resource and data source should include unit tests in a corresponding `*_test.go` file:
```bash
go test -v -count=1 ./internal/provider/...
```

### OpenAPI Coverage & Drift
Ensure that newly added endpoints maintain full OpenAPI coverage:
```bash
# Verify coverage mapping
go test -v ./tools/openapi/...
go run ./tools/openapi

# Check schema drift against live upstream
go run ./tools/openapi diff
```

### Local Live Testing Target
You can launch a local Jellyseerr instance using Docker Compose:
```bash
docker compose -f dev/docker-compose.yml up -d
```
Access the server at `http://localhost:5055` to test real provider configurations.

---

## 📝 Commit Conventions

| Prefix | Description | Example |
| :--- | :--- | :--- |
| `feat:` | Adds a new resource, data source, or major capability | `feat: add seerr_api_key resource (#162)` |
| `fix:` | Fixes a bug or schema discrepancy | `fix: prevent nil panic in notification read (#180)` |
| `docs:` | Updates documentation or guides | `docs: add compatibility guide (#182)` |
| `test:` | Adds or updates tests | `test: add contract tests for user quota (#175)` |
| `refactor:` | Code restructuring without schema changes | `refactor: extract HTTP client helper (#176)` |
| `ci:` | CI/CD workflows and automation | `ci: add openapi drift workflow (#182)` |

---

## 💬 Getting Help

Have questions or need guidance?
- Start a discussion in [GitHub Discussions](https://github.com/Josh-Archer/terraform-provider-seerr/discussions).
- Check the [Project Roadmap](ROADMAP.md) for planned phases and architecture decisions.
