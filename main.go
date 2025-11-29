package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	"tailscale.com/tsnet"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	hostname := os.Getenv("HOSTNAME")
	allowedClient := os.Getenv("ALLOWED_CLIENT")
	token := os.Getenv("TOKEN")

	if hostname == "" || allowedClient == "" || token == "" {
		log.Fatal("HOSTNAME, ALLOWED_CLIENT or TOKEN not set in .env")
	}

	srv := &tsnet.Server{
		Hostname: hostname,
	}
	defer srv.Close()

	ln, err := srv.Listen("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}

	lc, err := srv.LocalClient()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/lock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Token") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		who, err := lc.WhoIs(r.Context(), r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to determine identity", http.StatusInternalServerError)
			return
		}
		if who.Node.ComputedName != allowedClient {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		cmd := exec.Command("dms", "ipc", "call", "lock", "lock")
		if err := cmd.Run(); err != nil {
			http.Error(w, "failed to lock", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("pc-lock-service running on Tailnet HTTPS port 443")
	http.Serve(ln, nil)
}
