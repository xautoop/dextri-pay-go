package api

// Metadata is a JSON object containing non-sensitive, App-defined context.
// Values must be JSON-serializable and must never contain credentials or
// private wallet material.
type Metadata map[string]any
