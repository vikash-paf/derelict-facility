package assets

import "embed"

// AssetsFS embeds font and sound effect files used by the game engine into the compiled binary.
//go:embed fonts/FiraCodeNFBoldMono.ttf fonts/NotoEmoji-Regular.ttf audio/*.wav
var AssetsFS embed.FS
