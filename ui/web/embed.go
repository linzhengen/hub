package web

import "embed"

// Embedded contains embedded UI resources
//
//go:embed dist/*
var Embedded embed.FS
