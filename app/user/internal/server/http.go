package server

import (
	"net/http"

	"unionManageCenter/user/internal/service"
)

// New 构建并返回用户服务的 HTTP 路由
func New() http.Handler {
	mux := http.NewServeMux()

	svc := service.NewUserService()

	mux.HandleFunc("GET /api/v1/users", svc.List)
	mux.HandleFunc("POST /api/v1/users", svc.Create)
	mux.HandleFunc("GET /api/v1/users/{id}", svc.Get)
	mux.HandleFunc("PUT /api/v1/users/{id}", svc.Update)
	mux.HandleFunc("DELETE /api/v1/users/{id}", svc.Delete)

	return mux
}
