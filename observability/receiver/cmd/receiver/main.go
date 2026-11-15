package main

import (
	"log"
	"os"

	"receiver/internal/receiver"
)

func main() {
	cfg := receiver.LoadConfigFromEnv()
	addr := ":8081"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	srv := receiver.New(cfg)
	log.Printf("incident receiver listening on %s, bundle dir %s", addr, cfg.BundleRoot)
	if err := srv.ListenAndServe(addr); err != nil {
		log.Fatal(err)
	}
}
