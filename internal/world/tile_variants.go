package world

import "github.com/vikash-paf/derelict-facility/internal/core"

// Change this number if you ever add more tile types (like doors or water)
const maxTileTypes = 3

type TileAppearance struct {
	Char  string
	Color core.Color
}

type TileVariant [maxTileTypes]TileAppearance

// TileVariantClassic 1. Classic Rogue (The old-school standard)
var TileVariantClassic = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"#", core.White},
	TileTypeFloor: {".", core.White},
}

// TileVariantSolid 2. Heavy Concrete (Thick, solid walls with tiny floor dots)
var TileVariantSolid = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"█", core.White},
	TileTypeFloor: {"·", core.White},
}

// TileVariantGritty 3. Rusty & Gritty (Textured blocks, perfect for a derelict vibe)
var TileVariantGritty = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"▓", core.White},
	TileTypeFloor: {"░", core.White},
}

// TileVariantBlueprint 4. Clean Blueprint (Thin, structural walls)
var TileVariantBlueprint = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"╬", core.White},
	TileTypeFloor: {"◦", core.White},
}

// COLOR VARIANTS

// TileVariantToxic 5. Toxic Sector (Dark gray walls, glowing green floors)
var TileVariantToxic = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"█", core.Gray},
	TileTypeFloor: {"·", core.Green},
}

// TileVariantAlert 6. Red Alert (Harsh red walls, warning-yellow floors)
var TileVariantAlert = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"▓", core.Red},
	TileTypeFloor: {"░", core.Yellow},
}

// TileVariantCold 7. Cold Storage (Deep blue walls, icy cyan floors)
var TileVariantCold = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"█", core.Blue},
	TileTypeFloor: {".", core.Cyan},
}

// TileVariantHive 8. Alien Hive (Purple textured walls, green slime floors)
var TileVariantHive = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"▓", core.Magenta},
	TileTypeFloor: {"~", core.Green},
}

// TileVariantDark 9. Power Outage (Barely visible dark gray walls, faint white floors)
var TileVariantDark = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"#", core.DarkGray},
	TileTypeFloor: {".", core.White},
}

// todo: with a tile variants (theming) system implemented, now I can simulate weather, lightning, fire, etc.

var TileVariantLightning = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"█", core.BrightWhite}, // Bright white walls
	TileTypeFloor: {".", core.BrightWhite}, // Bright white floors
}

var TileVariantFlooded = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"█", core.Blue}, // Wet, dark blue walls
	TileTypeFloor: {"~", core.Cyan}, // Cyan ripples on the floor
}

var TileVariantAsh = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"▓", core.White},    // White frosted walls
	TileTypeFloor: {"*", core.DarkGray}, // Grey asterisks for ash/snow
}

var TileVariantPaused = TileVariant{
	TileTypeEmpty: {" ", core.Black},
	TileTypeWall:  {"▓", core.DarkGray}, // Dark Gray Wall
	TileTypeFloor: {"░", core.DarkGray}, // Dark Gray Floor
}
