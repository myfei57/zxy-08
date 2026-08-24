// Package web embeds the four console control pages.
package web

import _ "embed"

//go:embed topology.html
var TopologyHTML string

//go:embed keys.html
var KeysHTML string

//go:embed snapshots.html
var SnapshotsHTML string

//go:embed audit.html
var AuditHTML string
