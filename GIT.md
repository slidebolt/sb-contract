# Git Workflow for sb-contract

This repository contains the core Slidebolt contracts (interfaces and shared data structures) used across the entire workspace. It is a fundamental building block with zero internal dependencies.

## Dependencies
- **Internal:** None. This is a base repository.
- **External:** Standard Go library.

## Build Process
- **Type:** Pure Go Library (Shared Module).
- **Consumption:** Imported as a module dependency in other Go projects via `go.mod`.
- **Artifacts:** No standalone binary or executable is produced.
- **Validation:** 
  - Validated through unit tests: `go test -v ./...`
  - Validated by its consumers during their respective build/test cycles.

## Pre-requisites & Publishing
As a base repository, `sb-contract` is usually the **first** repository that needs to be updated and published when cross-cutting changes are made.

**Before publishing:**
1. Ensure all local tests pass.
2. If changing existing interfaces, check the impact on major consumers like `sb-api`, `sb-manager`, and all `plugin-*` repositories.

**Publishing Order:**
1. Update `sb-contract`.
2. Tag the repository (e.g., `git tag v1.0.3`).
3. Push the tag (`git push origin v1.0.3`).
4. Update dependent repositories using `go get github.com/slidebolt/sb-contract@v1.0.3`.

## Update Workflow & Verification
1. **Modify:** Update interfaces or data structures in `contract.go`.
2. **Verify Local:**
   - Run `go mod tidy`.
   - Run `go test ./...`.
3. **Verify Downstream:**
   - Since this is a core module, it is recommended to run a workspace-wide build if a breaking change is made.
   - Example: Run `go build ./...` in `sb-api` or a complex plugin to verify interface compliance.
4. **Commit:** Ensure the commit message clearly describes the contract change.
5. **Tag & Push:** (Follow the Publishing Order above).
