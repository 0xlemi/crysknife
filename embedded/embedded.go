package embedded

import "embed"

//go:embed templates/*
var Templates embed.FS

//go:embed hooks/*
var Hooks embed.FS
