package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config holds all configuration settings for spaceplant.
type Config struct {
	TileSizeW              int     `json:"tileSizeW"`              // Sprite width on the sprite sheet
	TileSizeH              int     `json:"tileSizeH"`              // Sprite height on the sprite sheet
	WorldWidth             int     `json:"worldWidth"`             // Draw width for the world portion of the screen
	WorldHeight            int     `json:"worldHeight"`            // Draw height for the world portion of the screen
	ScreenWidth            int     `json:"screenWidth"`            // Total window width
	ScreenHeight           int     `json:"screenHeight"`           // Total window height
	Title                  string  `json:"title"`                  // Window title
	BlueprintPath          string  `json:"blueprintPath"`          // Path to blueprint directory
	Lighting               bool    `json:"lighting"`               // Whether to render fog-of-war lighting
	ColorShading           bool    `json:"colorShading"`           // Whether to apply colour shading to sprites
	Los                    bool    `json:"los"`                    // Whether to apply line-of-sight culling
	PressDelay             int     `json:"pressDelay"`             // Key-repeat delay in ticks
	ShowMouseCoords        bool    `json:"showMouseCoords"`        // Overlay cursor coordinates
	RenderScale            float64 `json:"renderScale"`            // Viewport zoom: 1.0 = 1:1, 2.0 = each tile twice as large
	RenderPathfindingSteps bool    `json:"renderPathfindingSteps"` // Debug: draw AI pathfinding dots
	ProfileCPU             bool    `json:"profileCPU"`             // Whether to profile CPU usage
	ProfileMemory          bool    `json:"profileMemory"`          // Whether to profile memory usage
	DumpGenerationASCII    bool    `json:"dumpGenerationASCII"`    // Write gen_debug.txt after each generation
	NpcTurnDelayTicks      int     `json:"npcTurnDelayTicks"`      // Server ticks to pause between NPC-only turns (0 = no delay)
	CRTIntensity           float64 `json:"crtIntensity"`           // CRT shader intensity: 0.0 = off, 1.0 = full effect
}

// LoadConfig reads and unmarshals a JSON config file.
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
		return nil, err
	}
	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
		return nil, err
	}
	return &cfg, nil
}

var settings *Config

// Global returns the singleton config, loading it from data/config.json on first call.
func Global() *Config {
	if settings == nil {
		var err error
		settings, err = LoadConfig("data/config.json")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}
	return settings
}
