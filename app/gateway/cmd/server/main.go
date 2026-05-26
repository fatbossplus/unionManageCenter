package main

import (
	"fmt"
	"log"
	"os"

	"unionManageCenter/gateway/internal/router"
	"unionManageCenter/pkg/database"
)

func main() {
	err := database.Init(database.Config{
		Host:     envOrDefault("DB_HOST", "127.0.0.1"),
		Port:     3306,
		User:     envOrDefault("DB_USER", "root"),
		Password: envOrDefault("DB_PASS", "123456"),
		DBName:   envOrDefault("DB_NAME", "union_manage"),
	})
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	r := router.New()
	port := envOrDefault("PORT", "8080")
	log.Printf("server starting on :%s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
