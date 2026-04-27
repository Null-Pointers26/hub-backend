package main

import "sync"

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
	mu    sync.RWMutex
	games map[string]*Game // klíč = game ID
}

var registry = &Registry{games: make(map[string]*Game)}
var frontendTarget = "http://frontend:3000"
