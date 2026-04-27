package main

import (
	"encoding/json"
	"os"
	"sync"
)

type Game struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Target      string `json:"target"` // např. "http://game-chess:3000"
	Icon        string `json:"icon"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Image       string `json:"image"` // url na obrázek
}

type Registry struct {
	mu          sync.RWMutex
	games       map[string]*Game // klíč = game ID
	storagePath string
}

var registry = &Registry{games: make(map[string]*Game)}
var frontendTarget = "http://frontend:3000"

func (r *Registry) SaveToJSON() error {
	r.mu.RLock()
	data, err := json.MarshalIndent(r.games, "", "  ")
	r.mu.RUnlock()
	if err != nil {
		return err
	}

	return os.WriteFile(r.storagePath, data, 0o644)
}

func (r *Registry) LoadFromJSON() error {
	data, err := os.ReadFile(r.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	loadedGames := make(map[string]*Game)
	if err := json.Unmarshal(data, &loadedGames); err != nil {
		return err
	}

	r.mu.Lock()
	r.games = loadedGames
	r.mu.Unlock()

	return nil
}
