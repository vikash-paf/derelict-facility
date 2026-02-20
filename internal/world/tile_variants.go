package world

// Change this number if you ever add more tile types (like doors or water)
const maxTileTypes = 3

// ANSI Color Codes for easy mixing and matching
const (
	Reset   = "\x1b[0m"
	Red     = "\x1b[31m"
	Green   = "\x1b[32m"
	Yellow  = "\x1b[33m"
	Blue    = "\x1b[34m"
	Magenta = "\x1b[35m"
	Cyan    = "\x1b[36m"
	White   = "\x1b[37m"
	Gray    = "\x1b[90m"
)

// SIMPLE VARIANTS (No Colors)

// TileVariantClassic 1. Classic Rogue (The old-school standard)
var TileVariantClassic = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  "#",
	TileTypeFloor: ".",
}

// TileVariantSolid 2. Heavy Concrete (Thick, solid walls with tiny floor dots)
var TileVariantSolid = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  "█",
	TileTypeFloor: "·",
}

// TileVariantGritty 3. Rusty & Gritty (Textured blocks, perfect for a derelict vibe)
var TileVariantGritty = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  "▓",
	TileTypeFloor: "░",
}

// TileVariantBlueprint 4. Clean Blueprint (Thin, structural walls)
var TileVariantBlueprint = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  "╬",
	TileTypeFloor: "◦",
}

// COLOR VARIANTS

// TileVariantToxic 5. Toxic Sector (Dark gray walls, glowing green floors)
var TileVariantToxic = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  Gray + "█" + Reset,
	TileTypeFloor: Green + "·" + Reset,
}

// TileVariantAlert 6. Red Alert (Harsh red walls, warning-yellow floors)
var TileVariantAlert = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  Red + "▓" + Reset,
	TileTypeFloor: Yellow + "░" + Reset,
}

// TileVariantCold 7. Cold Storage (Deep blue walls, icy cyan floors)
var TileVariantCold = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  Blue + "█" + Reset,
	TileTypeFloor: Cyan + "." + Reset,
}

// TileVariantHive 8. Alien Hive (Purple textured walls, green slime floors)
var TileVariantHive = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  Magenta + "▓" + Reset,
	TileTypeFloor: Green + "~" + Reset,
}

// TileVariantDark 9. Power Outage (Barely visible dark gray walls, faint white floors)
var TileVariantDark = [maxTileTypes]string{
	TileTypeEmpty: " ",
	TileTypeWall:  Gray + "#" + Reset,
	TileTypeFloor: White + "." + Reset,
}
