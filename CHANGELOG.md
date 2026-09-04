# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
The project has not published its first version yet.

## Unreleased

### Added

- Compile-time compatibility checks for the planned `v0.1.0` public SDK surface.
- Validation tests for public request types and security-boundary tests for
  signing, retries, response limits, and webhook verification.
- CI coverage for the full quality gate and the minimum supported Go version.
- An opt-in Sandbox channel-discovery smoke test.
- App Commerce payment checkout, payment/refund queries, optional payouts, and
  payment/refund/payout webhook contracts for the planned `v0.1.0`.
- Balance responses expose `frozen` separately from authorization `locked`.
- Generic App-account balance, Hold create/get/release, multi-Hold commit, and
  atomic Escrow Settlement contracts. Settlement requests carry display and
  smallest-unit amounts and reject non-conserving allocations before transport.
- Hold, Escrow, and Settlement operation states and signed webhook snapshots.

### Changed

- Capability services now call the private transport directly through one
  executor boundary.
- CI uses portable Go cache defaults and current Node.js 24-based official
  Actions without enabling module caching for this dependency-free module.
- Stable request validation lives with public request types.
- Partner API requests use the module-specific `/pay/v1/` route prefix.
- Webhook event types, legacy compatibility, and verification are separated by
  responsibility.

### Removed

- The redundant private Resource/Provider forwarding layer.
