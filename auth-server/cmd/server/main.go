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
		organization   = pflag.String("organization", "", "GitHub organization name for team membership verification")
		team           = pflag.String("team", "", "GitHub team name for membership verification")
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

	// Validate team verification configuration
	if *team != "" && *organization == "" {
		log.Fatal("--organization is required when --team is specified for team membership verification")
	}

	log.Println("Starting GitHub App auth server...")
	if *verbose {
		log.Printf("Reading private key from: %s", *privateKeyPath)
	}

	// Log verification configuration
	if *team != "" && *organization != "" {
		log.Printf("Team membership verification enabled: organization=%s, team=%s", *organization, *team)
	} else if *organization != "" {
		log.Printf("Organization specified but no team - team membership verification disabled: organization=%s", *organization)
	} else {
		log.Println("No team membership verification configured")
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
		organization:  *organization,
		team:          *team,
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
	organization  string
	team          string
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

	// Get organization, team, and username from query parameters (optional if not configured on server)
	orgFromQuery := r.URL.Query().Get("org")
	teamFromQuery := r.URL.Query().Get("team")
	username := r.URL.Query().Get("username")

	// Use organization from query parameter if provided, otherwise use server configuration
	orgToCheck := h.organization
	if orgFromQuery != "" {
		orgToCheck = orgFromQuery
	}

	// Use team from query parameter if provided, otherwise use server configuration
	teamToCheck := h.team
	if teamFromQuery != "" {
		teamToCheck = teamFromQuery
	}

	// If server is configured with team or team is provided in query, require organization, team, and username
	if teamToCheck != "" {
		if orgToCheck == "" {
			if h.verbose {
				log.Printf("Organization is required for team verification but not provided from %s", r.RemoteAddr)
			}
			http.Error(w, "organization is required for team verification", http.StatusBadRequest)
			return
		}
		if username == "" {
			if h.verbose {
				log.Printf("Username is required for team verification but not provided from %s", r.RemoteAddr)
			}
			http.Error(w, "username query parameter is required for team verification", http.StatusBadRequest)
			return
		}
	}

	if h.verbose {
		log.Printf("Received token request for app-id=%s installation-id=%s org=%s team=%s username=%s from %s", 
			appID, installationID, orgToCheck, teamToCheck, username, r.RemoteAddr)
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

	// Perform team membership verification if organization, team, and username are provided
	if teamToCheck != "" && orgToCheck != "" && username != "" {
		if h.verbose {
			log.Printf("Verifying team membership for user %s in team %s of organization %s", username, teamToCheck, orgToCheck)
		}
		
		isMember, err := ghAuth.VerifyTeamMembership(token.Token, username, orgToCheck, teamToCheck)
		if err != nil {
			if h.verbose {
				log.Printf("Failed to verify team membership for user %s in %s/%s: %v", username, orgToCheck, teamToCheck, err)
			}
			http.Error(w, fmt.Sprintf("Failed to verify team membership: %v", err), http.StatusInternalServerError)
			return
		}

		if !isMember {
			if h.verbose {
				log.Printf("User %s is not a member of team %s in organization %s, denying token request", username, teamToCheck, orgToCheck)
			}
			http.Error(w, fmt.Sprintf("Access denied: user %s is not a member of team %s in organization %s", username, teamToCheck, orgToCheck), http.StatusForbidden)
			return
		}

		if h.verbose {
			log.Printf("User %s is a member of team %s in organization %s, allowing token request", username, teamToCheck, orgToCheck)
		}
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
