package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/CharlesSchwab/secrets-manager/auth-server/pkg/auth"
)

func main() {
	var (
		port          = flag.Int("port", 8080, "Port to listen on")
		privateKeyPath = flag.String("private-key-path", "", "Path to GitHub App private key PEM file")
	)
	flag.Parse()

	if *privateKeyPath == "" {
		log.Fatal("--private-key-path is required")
	}

	// Read private key file
	privateKeyPEM, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}

	handler := &Handler{
		privateKeyPEM: privateKeyPEM,
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
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func (h *Handler) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		http.Error(w, "app_id query parameter is required", http.StatusBadRequest)
		return
	}

	installationID := r.URL.Query().Get("installation_id")
	if installationID == "" {
		http.Error(w, "installation_id query parameter is required", http.StatusBadRequest)
		return
	}

	appIDInt, err := strconv.ParseInt(appID, 10, 64)
	if err != nil {
		http.Error(w, "invalid app_id", http.StatusBadRequest)
		return
	}

	instIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		http.Error(w, "invalid installation_id", http.StatusBadRequest)
		return
	}

	ghAuth, err := auth.NewGitHubAuth(h.privateKeyPEM, appIDInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize GitHub auth: %v", err), http.StatusInternalServerError)
		return
	}
	
	token, err := ghAuth.GetInstallationToken(instIDInt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get installation token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}