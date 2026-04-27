package main

import (
	"encoding/json"
	"net/http"
)

func addGameHandler(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	registry.mu.Lock()
	registry.games[game.ID] = &game
	registry.mu.Unlock()

	if err := registry.SaveToJSON(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

func removeGameHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	registry.mu.Lock()
	delete(registry.games, id)
	registry.mu.Unlock()

	if err := registry.SaveToJSON(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func listGamesHandler(w http.ResponseWriter, r *http.Request) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	games := make([]*Game, 0, len(registry.games))
	for _, g := range registry.games {
		games = append(games, g)
	}

	json.NewEncoder(w).Encode(games)
}
