# Map and Mission Specification — Derelict Facility (v0.5.0)

This document defines the constraints, formats, symbol-item maps, and engine mechanics required to construct custom maps and missions for the Derelict Facility engine. It serves as a comprehensive reference for automated systems or content creators.

---

## 1. Directory Structure Conventions

Missions can reside in two locations:

1. **Embedded / Dev Repository Location**: Located under the `assets/missions/` directory. Useful for built-in or pre-bundled story campaigns.
2. **External / Binary Parallel Location**: Located under the `missions/` directory parallel to the executable binary (e.g. `./missions/`). Useful for user-created custom levels and distribution.

Each mission directory contains a metadata file and a subfolder for its levels:

```
[missions_root]/[mission_id]/
├── mission.json
└── maps/
    ├── [level_01_id].json
    ├── [level_02_id].json
    └── ...
```

---

## 2. Mission Metadata Schema (mission.json)

The `mission.json` file defines the entry point, level sequence, and player-facing description of the campaign.

### Schema Fields
- `id` (string): Unique identifier for the mission. Matches the directory name.
- `title` (string): User-facing mission title shown in the main menu.
- `author` (string): Creator's name or alias.
- `synopsis` (string): A short overview of the mission objectives.
- `start_level` (string): The ID of the starting level.
- `levels` (array): Sequence of maps defining the campaign progression.
  - `id` (string): Unique identifier for the level.
  - `name` (string): User-facing name of the level.
  - `file` (string): Relative path from the mission folder to the level JSON map file.

### Example mission.json
```json
{
  "id": "sector_4_incident",
  "title": "MISSION 1: SECTOR 4 CONTAINMENT BREACH",
  "author": "System Architect",
  "synopsis": "Investigate Sector 4, descend into Sub-Level Labs, and restore the Deep Core Reactor.",
  "start_level": "level_01_surface",
  "levels": [
    {
      "id": "level_01_surface",
      "name": "Surface Hangar Bay & Sunlit Skylight Corridor",
      "file": "maps/level_01_surface.json"
    },
    {
      "id": "level_02_labs",
      "name": "Sub-Level Research Labs & Specimen Vault",
      "file": "maps/level_02_labs.json"
    },
    {
      "id": "level_03_reactor",
      "name": "Deep Subcore Reactor & Turbine Complex",
      "file": "maps/level_03_reactor.json"
    }
  ]
}
```

---

## 3. Map Data Schema and Spatial Constraints

Maps are defined in JSON format and consist of a grid of characters that the engine parses into tiles and ECS entities.

### Structural Requirements
- `width` (integer): Must be exactly 120.
- `height` (integer): Must be exactly 40.
- `rows` (array of strings): Must contain exactly 40 strings, each exactly 120 characters long.

### Architectural Constraints
- Outer Boundary: Every map must be surrounded by a solid perimeter of walls (`#`) to prevent players from going out of bounds.
- Shared Walls: Adjacent rooms must share a single wall border. Do not construct double-thick wall buffers between adjacent rooms.
- Single Doorways: Connections between rooms must be exactly one tile wide (`+`). Double-width door gaps are not supported by the autotiler and pathfinder.

---

## 4. Map Character Symbol Mappings

When the JSON map is parsed, characters in the `rows` grid are converted to base tiles and populated with ECS components:

| Character | Tile Type | Walkable | Collidable | Spawned Components / Actions |
|-----------|-----------|----------|------------|------------------------------|
| `#`       | Wall      | No       | Yes        | Base wall tile. Renders with automatic corner/intersection bitmask tiling. |
| `.`       | Floor     | Yes      | No         | Walkable floor tile. |
| `S` / `*` | Floor     | Yes      | No         | Sunlit Skylight floor tile. Walkable and exposed to the dynamic solar cycle. |
| `+`       | Floor     | Yes      | Yes (init) | Door entity. Starts closed (`+`, solid). Opens to (`/`, non-solid) on interaction. |
| `@`       | Floor     | Yes      | No         | Player start location. Defines default coordinate if no elevator transition occurs. |
| `>`       | Floor     | Yes      | No         | Descending elevator/stairway transition. Moves player to the next level in sequence. |
| `<`       | Floor     | Yes      | No         | Ascending elevator/stairway transition. Moves player to the previous level in sequence. |
| `T`       | Floor     | Yes      | Yes        | Interactive save terminal. Interacting triggers game saving. |
| `G`       | Floor     | Yes      | Yes        | Global Power Generator. Powering it turns on lights across the entire map. |
| `g`       | Floor     | Yes      | Yes        | Local Room Power Generator. Powering it turns on lights only in its local room bounds. |

---

## 5. Core Game Mechanics and Systems

### 5.1 Power Grid and Visibility
- Powered Rooms: When a room's generator is inactive, the room remains dark. The player must rely on their direct field-of-view (FOV) radius of 8 tiles.
- Powered State: If a room generator (`g`) is toggled on, it illuminates its local derived room boundary. If the global generator (`G`) is toggled on, it illuminates the entire map layout.
- Visibility: All tiles in powered rooms or a powered map are set to `Visible` and `Explored` automatically, bypassing the FOV raycasting limits.
- Solar Skylight Cycle: Tiles marked with `S` or `*` receive varying light levels depending on the active facility clock, transitioning through day, dusk, night, and dawn.

### 5.2 Room Derivation Algorithm
- The engine uses a flood-fill (Breadth-First Search) algorithm on walkable floors to dynamically determine room boundaries.
- Doorways (`+`) act as partition boundaries. A room is defined as a contiguous set of walkable floor tiles surrounded by doors or walls.
- Local generators (`g`) calculate which room they belong to based on their spatial grid coordinate matching a derived room's bounding box.

### 5.3 Level Sequence and Transitions
- Transition Triggers: Standing on or adjacent to an elevator symbol (`>` or `<`) and pressing the interaction key (`[E]`) triggers a level swap.
- SEQUENCE LOGIC:
  - Descending (`>`) looks up the current level index in `mission.json` and loads the level at index `current + 1`.
  - Ascending (`<`) looks up the current level index in `mission.json` and loads the level at index `current - 1`.
- Spawn Positioning:
  - If a player descends via `>` in the current level, they arrive at the corresponding ascending elevator `<` in the next level.
  - If a player ascends via `<` in the current level, they arrive at the corresponding descending elevator `>` in the previous level.
  - The player spawns exactly 1 tile to the right of the destination elevator (`x + 1`, `y`) to avoid colliding with the stairway tile. If the target elevator is not found, the player spawns at the fallback `@` coordinates.

### 5.4 Door Security Clearance
- Doors can be locked by setting a `RequiredClearance` mask value (bitmask).
- When a player interacts with a door, the engine performs a bitwise AND between the player's held `SecurityClearance` mask and the door's `RequiredClearance` value.
- If the player's clearance mask has the matching bits set, the door opens. If not, the engine plays an access denied sound effect.
- Keycards and terminal access events update the player's `SecurityClearance` bitmask.
