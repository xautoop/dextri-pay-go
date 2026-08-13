# Release checklist

This repository is preparing for its first `v0.1.0` release. Publishing a tag
is a separate, explicit operation and is not part of ordinary development.

1. Confirm `CHANGELOG.md` describes every public behavior change.
2. Run `make check` with the declared Go toolchain.
3. Run `make test-sandbox` with dedicated Sandbox credentials.
4. Review the compile-time API manifest in
   `internal/architecture/public_api_test.go`.
5. Verify installation from a clean temporary module using the candidate
   commit or tag.
6. Create and push the signed version tag only after release approval.
7. Verify the published module and pkg.go.dev documentation.

Never use Production credentials for the Sandbox smoke test or store secrets in
the repository, test output, release notes, or CI configuration.
