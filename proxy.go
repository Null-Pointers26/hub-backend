package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func buildProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(u), nil
}

func routerHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/games/") {
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/games/"), "/", 2)
		gameID := parts[0]

		registry.mu.RLock()
		game, ok := registry.games[gameID]
		registry.mu.RUnlock()

		if !ok {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		proxy, err := buildProxy(game.Target)
		if err != nil {
			http.Error(w, "Bad gateway", http.StatusBadGateway)
			return
		}

		if len(parts) > 1 {
			r.URL.Path = "/" + parts[1]
		} else {
			r.URL.Path = "/"
		}

		proxy.ServeHTTP(w, r)
		return
	}

	proxy, err := buildProxy(frontendTarget)
	if err != nil {
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(w, r)
}
