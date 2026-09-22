package main

import "embed"

// Embedded shipped defaults.
//
// skills/ and souls/ live at the repository root, and Go's embed directive
// cannot reference parent directories — so the FS has to be declared here in
// package main and handed to the layout seeder.
//
// This matters for packaged builds: the installer does not place skills/ or
// souls/ next to the executable, so without embedding a fresh install would
// give the agent no skill playbooks and only a generic fallback persona.
//
//go:embed all:skills all:souls
var defaultsFS embed.FS
