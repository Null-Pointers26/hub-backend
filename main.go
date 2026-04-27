package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
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
