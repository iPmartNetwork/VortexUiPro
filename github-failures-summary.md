# GitHub Actions Failures - Exact Error Annotations

## Summary
All failing jobs on the `master` branch for commit `cb1081affb811b68d0fbea6dde7bfe6929b7659e` share a common root cause: **Docker build failures with `go mod download` exit code 1** for the node and panel images, plus a **lint/test failure** in the CI workflow.

---

## 1. CI / Lint (Job ID: 90822613523)

**Workflow**: CI  
**Job**: Lint (Go + Frontend)  
**Status**: failure

### Annotations:
1. **Warning** (line 2):
   - Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: actions/setup-go@v5.

2. **Failure** (line 12):
   ```
   Process completed with exit code 1.
   ```

---

## 2. Docker Build (node) - from Docker Build workflow (Job ID: 90822613515)

**Workflow**: Docker Build  
**Job**: Build Docker Images (node, deploy/Dockerfile, node, vortexuipro-node)  
**Status**: failure

### Annotations:
1. **Warning** (line 2):
   - Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: actions/upload-artifact@v5, docker/build-push-action@v6, docker/setup-buildx-action@v3.

2. **Warning** (line 13):
   ```
   No files were found with the provided path: trivy-node.sarif. No artifacts will be uploaded.
   ```

3. **Failure** (line 299):
   ```
   buildx failed with: ERROR: failed to build: failed to solve: process "/bin/sh -c go mod download" did not complete successfully: exit code: 1
   ```

---

## 3. Docker Build (panel) - from Docker Build workflow (Job ID: 90822613471)

**Workflow**: Docker Build  
**Job**: Build Docker Images (panel, deploy/Dockerfile, panel, vortexuipro-panel)  
**Status**: failure

### Annotations:
1. **Warning** (line 2):
   - Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: actions/upload-artifact@v5, docker/build-push-action@v6, docker/setup-buildx-action@v3.

2. **Warning** (line 13):
   ```
   No files were found with the provided path: trivy-panel.sarif. No artifacts will be uploaded.
   ```

3. **Failure** (line 392):
   ```
   buildx failed with: ERROR: failed to build: failed to solve: process "/bin/sh -c go mod download" did not complete successfully: exit code: 1
   ```

---

## 4. Auto Release / Build & Release (Job ID: 90822686029)

**Workflow**: 🚀 Auto Release  
**Job**: 📦 Build & Release  
**Status**: failure

### Annotations:
1. **Warning** (line 5):
   - Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: actions/setup-go@v5.

2. **Failure** (line 47):
   ```
   Process completed with exit code 1.
   ```

---

## 5. Auto Deploy / Docker Build (node) (Job ID: 90822734877)

**Workflow**: 🚀 Auto Deploy  
**Job**: 🐳 Docker Build (node, deploy/Dockerfile, node)  
**Status**: failure

### Annotations:
1. **Warning** (line 2):
   - Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: docker/build-push-action@v6, docker/login-action@v3, docker/setup-buildx-action@v3.

2. **Failure** (line 337):
   ```
   buildx failed with: ERROR: failed to build: failed to solve: process "/bin/sh -c go mod download" did not complete successfully: exit code: 1
   ```

---

## 6. Auto Deploy / Docker Build (panel) (Job ID: 90822734907)

**Workflow**: 🚀 Auto Deploy  
**Job**: 🐳 Docker Build (panel, deploy/Dockerfile, panel)  
**Status**: failure

### Annotations:
1. **Warning** (line 2):
   - Node.js 20 is deprecated. The following actions target Node.js 20 but are being forced to run on Node.js 24: docker/build-push-action@v6, docker/login-action@v3, docker/setup-buildx-action@v3.

2. **Failure** (line 476):
   ```
   buildx failed with: ERROR: failed to build: failed to solve: process "/bin/sh -c go mod download" did not complete successfully: exit code: 1
   ```

---

## 7. Auto Deploy / Validate (Job ID: 90822735003)

**Workflow**: 🚀 Auto Deploy  
**Job**: 🔍 Validate  
**Status**: skipped (no annotations, skipped due to upstream failures)

---

## Root Causes

### Primary Issue: `go mod download` failure
All Docker builds for `node` and `panel` targets fail at the same step:
```
process "/bin/sh -c go mod download" did not complete successfully: exit code: 1
```

This indicates a Go module dependency resolution issue during the Docker build process. The commit message mentions "use pure-Go SQLite for CGO-free multi-arch panel build", suggesting the recent changes to dependencies (likely around SQLite) are causing the `go mod download` to fail.

### Secondary Issue: CI Lint failure
The CI workflow's lint job fails with exit code 1 (line 12), though the exact command output isn't in the annotations. This could be related to:
- Go lint errors introduced by the SQLite dependency changes
- Frontend lint errors
- Test failures that run during the lint step

### Warnings (non-blocking):
- Node.js 20 deprecation warnings across all workflows
- Missing Trivy SARIF artifacts (security scan outputs not generated due to build failures)

---

## Recommended Actions

1. **Investigate `go mod download` failure**: Check the actual build logs for the Docker jobs to see which specific Go module is failing to download. Look at:
   - Changes to `go.mod` / `go.sum` in commit `cb1081affb811b68d0fbea6dde7bfe6929b7659e`
   - Network issues or proxy settings in the Docker build environment
   - Module checksums or unavailable dependencies

2. **Fix CI lint failure**: Run the lint job locally or check the full CI logs to identify the specific linting/test error.

3. **Address Node.js deprecation**: Update GitHub Actions to use Node.js 24-compatible versions (though this is a warning, not blocking).

4. **Verify SQLite changes**: The commit message mentions switching to pure-Go SQLite. Ensure:
   - The new SQLite library is correctly added to `go.mod`
   - Build tags or CGO settings are properly configured
   - The dependency is available in the Docker build context
