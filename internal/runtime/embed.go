package runtime

import _ "embed"

// SourceServerGo is the source code of server.go, embedded at build time.
// Used by internal/build to push the standalone runtime into ephemeral
// build repositories. The contents are a snapshot at compile time and
// always match server.go exactly.
//
//go:embed server.go
var SourceServerGo string

// SourceStaticGo is the source code of static.go, embedded at build time.
// Used by internal/build to push the standalone runtime into ephemeral
// build repositories. The contents are a snapshot at compile time and
// always match static.go exactly.
//
//go:embed static.go
var SourceStaticGo string
