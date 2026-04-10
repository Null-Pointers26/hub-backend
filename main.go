package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/acme/autocert"
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
	mu    sync.RWMutex
	games map[string]*Game // klíč = game ID
}

var registry = &Registry{games: make(map[string]*Game)}
var frontendTarget = "http://frontend:3000"

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// --- Proxy logika ---

func buildProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(u), nil
}

// routing podle prefixu cesty
func routerHandler(w http.ResponseWriter, r *http.Request) {
	// /games/{id}/... → proxy na konkrétní hru
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

		// předej cestu za ID hry, nebo "/" pokud nic není
		if len(parts) > 1 {
			r.URL.Path = "/" + parts[1]
		} else {
			r.URL.Path = "/"
		}

		proxy.ServeHTTP(w, r)
		return
	}

	// přesměrování na frontend
	proxy, err := buildProxy(frontendTarget)
	if err != nil {
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		return
	}
	proxy.ServeHTTP(w, r)
}

// správa her

func addGameHandler(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	registry.mu.Lock()
	registry.games[game.ID] = &game
	registry.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

func removeGameHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	registry.mu.Lock()
	delete(registry.games, id)
	registry.mu.Unlock()
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

func main() {
	domain := getEnv("DOMAIN", "localhost")
	certEmail := os.Getenv("CERT_EMAIL")
	certCacheDir := getEnv("CERT_CACHE_DIR", "/certs")
	frontendTarget = getEnv("FRONTEND_TARGET", "http://frontend:3000")
	appEnv := getEnv("APP_ENV", "development")
	devListenAddr := getEnv("DEV_LISTEN_ADDR", ":8080")
	httpAddr := getEnv("HTTP_ADDR", ":80")
	httpsAddr := getEnv("HTTPS_ADDR", ":443")

	mux := http.NewServeMux()

	// Veřejné API — seznam her (pro frontend)
	mux.HandleFunc("GET /api/games", listGamesHandler)

	// Chráněné API — správa her

	// Proxy router — / a /games/
	mux.HandleFunc("/", routerHandler)

	// --- HTTPS s autocert ---
	certManager := &autocert.Manager{
		Cache:      autocert.DirCache(certCacheDir), // persistentní volume
		Prompt:     autocert.AcceptTOS,
		Email:      certEmail,
		HostPolicy: autocert.HostWhitelist(domain),
	}

	httpsServer := &http.Server{
		Addr:    httpsAddr,
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	if appEnv == "production" {
		fmt.Printf("Starting hub backend in production mode (domain: %s)\n", domain)
		fmt.Printf("HTTP redirect/challenge server listening on %s\n", httpAddr)
		fmt.Printf("HTTPS server listening on %s\n", httpsAddr)

		// Port 80 — pouze redirect na HTTPS + ACME challenge
		go func() {
			httpMux := certManager.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				target := "https://" + r.Host + r.RequestURI
				http.Redirect(w, r, target, http.StatusMovedPermanently)
			}))
			if err := http.ListenAndServe(httpAddr, httpMux); err != nil && err != http.ErrServerClosed {
				fmt.Println("HTTP redirect server error:", err)
			}
		}()

		// autocert HTTPS
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			fmt.Println("HTTPS server error:", err)
		}
	} else {
		fmt.Printf("Starting hub backend in development mode on %s\n", devListenAddr)

		// Lokálně bez TLS
		if err := http.ListenAndServe(devListenAddr, mux); err != nil && err != http.ErrServerClosed {
			fmt.Println("HTTP server error:", err)
		}
	}
}
