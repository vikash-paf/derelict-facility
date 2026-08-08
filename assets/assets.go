package assets

import "embed"

// AssetsFS embeds only the font files used by the game engine into the compiled binary.
//go:embed fonts/FiraCodeNFBoldMono.ttf fonts/NotoEmoji-Regular.ttf
var AssetsFS embed.FS
