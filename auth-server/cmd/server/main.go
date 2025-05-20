package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gclhub/gh-secrets-manager/auth-server/pkg/auth"
	"github.com/spf13/pflag"
)

func main() {
	var (
		port           = pflag.Int("port", 8080, "Port to listen on")
		privateKeyPath = pflag.String("private-key-path", "", "Path to GitHub App private key PEM file")
		verbose        = pflag.BoolP("verbose", "v", false, "Enable verbose logging")
		help           = pflag.BoolP("help", "h", false, "Show help message")
	)
	
	// Add support for --flag syntax
	pflag.CommandLine.SetNormalizeFunc(pflag.CommandLine.GetNormalizeFunc())
	pflag.Parse()
	
	// Handle help flag manually to avoid "pflag: help requested" message
	if *help {
		pflag.Usage()
		os.Exit(0)
	}
	
	// Set verbosity level
	auth.Verbose = *verbose

	if *privateKeyPath == "" {
		log.Fatal("--private-key-path is required")
	}

	log.Println("Starting GitHub App auth server...")
	if *verbose {
		log.Printf("Reading private key from: %s", *privateKeyPath)
	}

	// Read private key file
	privateKeyPEM, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}
	if *verbose {
		log.Printf("Successfully loaded private key")
	}

	handler := &Handler{
		privateKeyPEM: privateKeyPEM,
		verbose:       *verbose,
	}

	http.HandleFunc("/healthz", handler.handleHealth)
	http.HandleFunc("/token", handler.handleToken)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

type Handler struct {
	privateKeyPEM []byte
	verbose       bool
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func (h *Handler) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		if h.verbose {
			log.Printf("Invalid method %s from %s", r.Method, r.RemoteAddr)
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("app-id")
	if appID == "" {
		if h.verbose {
			log.Printf("Missing app-id parameter from %s", r.RemoteAddr)
		}
		http.Error(w, "app-id query parameter is required", http.StatusBadRequest)
		return
	}

	installationID := r.URL.Query().Get("installation-id")
	if installationID == "" {
		if h.verbose {
			log.Printf("Missing installation-id parameter from %s", r.RemoteAddr)
		}
		http.Error(w, "installation-id query parameter is required", http.StatusBadRequest)
		return
	}

	if h.verbose {
		log.Printf("Received token request for app-id=%s installation-id=%s from %s", appID, installationID, r.RemoteAddr)
	}

	appIDInt, err := strconv.ParseInt(appID, 10, 64)
	if err != nil {
		if h.verbose {
			log.Printf("Invalid app-id %s from %s: %v", appID, r.RemoteAddr, err)
		}
		http.Error(w, "invalid app-id", http.StatusBadRequest)
		return
	}

	instIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		if h.verbose {
			log.Printf("Invalid installation-id %s from %s: %v", installationID, r.RemoteAddr, err)
		}
		http.Error(w, "invalid installation-id", http.StatusBadRequest)
		return
	}

	ghAuth, err := auth.NewGitHubAuth(h.privateKeyPEM, appIDInt)
	if err != nil {
		if h.verbose {
			log.Printf("Failed to initialize GitHub auth for app-id=%s from %s: %v", appID, r.RemoteAddr, err)
		}
		http.Error(w, fmt.Sprintf("Failed to initialize GitHub auth: %v", err), http.StatusInternalServerError)
		return
	}

	if h.verbose {
		log.Printf("Initialized GitHub auth successfully, requesting installation token...")
		log.Printf("Getting installation token for app-id=%s installation-id=%s", appID, installationID)
	}
	token, err := ghAuth.GetInstallationToken(instIDInt)
	if err != nil {
		if h.verbose {
			log.Printf("Failed to get installation token for app-id=%s installation-id=%s: %v", appID, installationID, err)
		}
		http.Error(w, fmt.Sprintf("Failed to get installation token: %v", err), http.StatusInternalServerError)
		return
	}

	if h.verbose {
		log.Printf("Successfully generated token for app-id=%s installation-id=%s valid until %s", appID, installationID, token.ExpiresAt)
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(token); err != nil {
		if h.verbose {
			log.Printf("Failed to encode token response: %v", err)
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if h.verbose {
		log.Printf("Successfully sent token response to client %s", r.RemoteAddr)
	}
}
