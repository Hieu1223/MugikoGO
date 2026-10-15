package routes

import (
	"net/http"

	"mugiko/internal/modules/core/user/api/routes/handlers"
	"mugiko/internal/modules/core/user/middleware"
)

func RegisterUserRoutes(mux *http.ServeMux) {
	// POST /api/v1/users/register
	mux.HandleFunc("POST /api/v1/users/register", handlers.RegisterUser)

	// GET /api/v1/users/me - requires auth
	mux.Handle("GET /api/v1/users/me", middleware.RequireUser(http.HandlerFunc(handlers.GetProfile)))

	// PATCH /api/v1/users/me/profile - requires auth
	mux.Handle("PATCH /api/v1/users/me/profile", middleware.RequireUser(http.HandlerFunc(handlers.UpdateProfile)))
}

func RegisterAdminRoutes(mux *http.ServeMux) {
	// GET /api/v1/admin/users - requires admin
	mux.Handle("GET /api/v1/admin/users", middleware.RequireAdmin(http.HandlerFunc(handlers.ListUsers)))

	// GET /api/v1/admin/users/{user_id} - requires admin
	mux.Handle("GET /api/v1/admin/users/{user_id}", middleware.RequireAdmin(http.HandlerFunc(handlers.GetUserDetail)))

	// PATCH /api/v1/admin/users/{user_id}/role - requires admin
	mux.Handle("PATCH /api/v1/admin/users/{user_id}/role", middleware.RequireAdmin(http.HandlerFunc(handlers.UpdateUserRole)))

	// PATCH /api/v1/admin/users/{user_id}/deactivate - requires admin
	mux.Handle("PATCH /api/v1/admin/users/{user_id}/deactivate", middleware.RequireAdmin(http.HandlerFunc(handlers.DeactivateUser)))
}