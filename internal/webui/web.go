package webui

import (
	"embed"
	"io/fs"
)

// Files contains the static dashboard assets embedded in the Polaris binary.
//
//go:embed index.html assets/*
var embedded embed.FS

func Files() fs.FS {
	return embedded
}
