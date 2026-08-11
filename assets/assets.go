package assets

import "embed"

// AssetsFS embeds fonts, sound effects, and story campaign missions into the compiled binary.
//
//go:embed fonts/FiraCodeNFBoldMono.ttf fonts/NotoEmoji-Regular.ttf audio/*.wav missions/*
var AssetsFS embed.FS
