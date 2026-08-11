# Release Notes (v0.6.0)

## Key Features and Gameplay Systems
* Survival and Hydroponics Systems:
  * Introduced player survival mechanics (oxygen, health, and life support) and hydroponics harvesting systems.
  * Added PlayerSurvival and Hydroponics ECS components.
  * Seeded map tiles with harvestable Hydroponics (H) plants.
  * Refactored plants from dynamic ECS entities to static map tiles to optimize performance and prevent entity overflow.
* Survival Balancing:
  * Implemented passive health regeneration under light sources and tuned medpack efficacy.
  * Balanced oxygen depletion, recovery rates, and passive health regeneration rates inside safe zones.
* Persistent Level Transitions:
  * Added transition handling to maintain player survival state metrics across different levels.
  * Implemented a level cache to persist level states during transitions and exported level saving callbacks.
  * Saved the active mission state alongside game saves.

## Visuals and Audio FX
* Immersive Effects:
  * Added procedurally synthesized heartbeat audio (heartbeat.wav).
  * Implemented dynamic heartbeat warning tempos matching low-health thresholds.
  * Added sickness camera sways and emergency lighting flickering effects.
* Game Over Screen:
  * Added an interactive Game Over overlay screen allowing players to restore from checkpoints or return to the main menu.
  * Exposed descriptive error feedback on checkpoint restoration failures.
* HUD Improvements:
  * Adjusted HUD metric bar spacing to prevent text overlaps.
  * Added support for 4K font scaling.
  * Displayed active mission titles directly on the HUD.

## Missions and Maps
* New Mission and Map Content:
  * Added a new mission: "biodome collapse".
  * Added support for "Protocol 9" mission.
  * Grouped custom missions into a dedicated subdirectory.
* Map and Tech Specs:
  * Added comprehensive map generation specifications and developer documentation.
  * Fixed biodome map loading bugs and path issues.

## Development and codebase
* Raylib Integration:
  * Migrated legacy custom rendering and windowing code to utilize native Raylib features.
  * Fixed player depth rendering (z-index) to draw the player above open door outlines.
  * Disabled custom shader steps.
* Build System:
  * Added a Makefile supporting run, build, and rerun targets.
