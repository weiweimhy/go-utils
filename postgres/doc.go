// Package postgres provides a small pgx connection pool constructor.
//
// NewPool validates its configuration, creates a pgxpool.Pool, and verifies
// connectivity with Ping before returning it. StartupTimeout defaults to ten
// seconds and covers only pool creation and the startup Ping. It does not set a
// timeout for later queries.
//
// Pool connection limits and related settings can be supplied with pgxpool's
// native pool_* connection string parameters. An empty connection string uses
// pgx's normal environment-variable behavior. NewPool rejects a nonpositive
// pool health check period, negative minimum connection counts, and minimum
// counts greater than the configured maximum.
//
// The returned pool is the native pgxpool.Pool. Create it once for the
// application lifecycle and reuse it across requests. Pool.Close has no context
// or return value; it waits synchronously for checked-out connections and
// transactions to be released and can be called repeatedly. Always close Rows
// returned by queries so the pool can reclaim the connection.
//
// Errors returned by NewPool have fixed, sanitized Error text and preserve the
// underlying error chain for errors.Is and errors.As. Inspecting an unwrapped
// driver error may reveal connection details, including credentials from the
// connection string. This module requires Go 1.27.2 or newer.
package postgres
