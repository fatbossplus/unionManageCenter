package main

import (
	"log"
	"net/http"

	"unionManageCenter/user/internal/server"
)

func main() {
	srv := server.New()
	log.Println("user-service starting on :8081")
	if err := http.ListenAndServe(":8081", srv); err != nil {
		log.Fatalf("user-service failed: %v", err)
	}
}
