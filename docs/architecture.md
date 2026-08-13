# SDK Architecture

Dextri Pay Go uses one explicit client facade, capability-oriented public
packages, and small private infrastructure packages.

## Dependency direction

```text
merchant application
        |
        v
client.Client and capability services
        |                         |
        v                         v
public request/response types   internal/transport
                                  |       |
                                  v       v
                            internal/auth internal/retry
```

`client.Client` constructs every capability service once. A service coordinates
stable request validation with endpoint mapping and response decoding, and calls
the private transport through the `executor` interface. There is intentionally
no second resource/provider forwarding layer.

Public domain packages contain API value types and stable validation rules.
They do not import private implementation packages. The transport is domain
agnostic and must not import capability packages.

Webhook verification is independent of the API client. Event types, legacy
payload compatibility, and signature verification are separated by file while
remaining in one public `webhook` package.

Architecture tests under `internal/architecture` enforce these boundaries.
