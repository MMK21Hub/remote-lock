package main

import (
	"log"
	"net/http"
	"os/exec"

	"tailscale.com/tsnet"
	"tailscale.com/types/netmap"
)

const allowedClient = "home-assistant"   // change to your HA device's Tailscale name
const token = "changeme-superlong-token" // put the same in HA

func main() {
	// Create a tsnet node inside your Tailnet
	srv := &tsnet.Server{
		Hostname: "remote-lock-service",
	}

	// Starts a TLS listener on port 443 inside Tailnet
	ln, err := srv.Listen("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}

	// Single route
	http.HandleFunc("/lock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}

		// Token check
		if r.Header.Get("X-Token") != token {
			http.Error(w, "unauthorized", 401)
			return
		}

		// Tailscale identity check
		who, ok := netmap.TailscaleIdentityFromRequest(r)
		if !ok || who.Node.Name() != allowedClient {
			http.Error(w, "forbidden", 403)
			return
		}

		// Execute the lock command
		cmd := exec.Command("dms", "ipc", "call", "lock", "lock")
		if err := cmd.Run(); err != nil {
			http.Error(w, "failed to lock", 500)
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	log.Println("remote-lock running on Tailnet HTTPS port 443")
	http.Serve(ln, nil)
}
