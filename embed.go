package main

import "embed"

//go:embed systems/*.json
var bundledSystems embed.FS
