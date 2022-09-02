package kitsh

import "embed"

// NoVNCEmbed is an embedded noVNC application for viewing VNC displays.
//go:embed noVNC/vnc_lite.html noVNC/core/* noVNC/vendor/*
var NoVNCEmbed embed.FS
