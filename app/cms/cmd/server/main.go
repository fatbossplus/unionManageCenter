package main

import (
	"fmt"
	"log"
	"os"

	"unionManageCenter/cms/internal/router"
	"unionManageCenter/cms/internal/worker"
	"unionManageCenter/pkg/database"

	// 注册所有适配器（import 副作用）
	_ "unionManageCenter/cms/internal/adapter/scraper"
	_ "unionManageCenter/cms/internal/adapter/ai"
	_ "unionManageCenter/cms/internal/adapter/compliance"
	_ "unionManageCenter/cms/internal/adapter/publisher"
)

func main() {
	// 初始化数据库
	err := database.Init(database.Config{
		Host:     envOr("DB_HOST", "127.0.0.1"),
		Port:     3306,
		User:     envOr("DB_USER", "root"),
		Password: envOr("DB_PASS", "123456"),
		DBName:   envOr("DB_NAME", "union_manage"),
	})
	if err != nil {
		log.Fatalf("CMS: 数据库初始化失败: %v", err)
	}
	log.Println("CMS: 数据库连接成功")

	// 启动后台工作器
	scheduler := worker.NewScheduler()
	scheduler.Start()
	defer scheduler.Stop()

	processor := worker.NewProcessor()
	processor.Start()
	defer processor.Stop()

	// 启动 HTTP 服务
	r := router.New()
	port := envOr("CMS_PORT", "8082")
	log.Printf("CMS 服务启动，监听端口 :%s", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("CMS: 服务启动失败: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
