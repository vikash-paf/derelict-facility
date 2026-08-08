package assets

import "embed"

// AssetsFS embeds only the fonts and tilesets actively used by the game engine into the compiled binary.
//go:embed fonts/FiraCodeNFBoldMono.ttf fonts/NotoEmoji-Regular.ttf tileset.png
var AssetsFS embed.FS
