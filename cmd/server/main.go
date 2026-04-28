package main

import (
	"fmt"
	"log"
	"net/http"

	"field_service_management_api/internal/config"
	"field_service_management_api/internal/db"
	"field_service_management_api/internal/handlers"
	"field_service_management_api/internal/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present (silently ignore if missing in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("Database connected")

	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/auth/register", handlers.Register(pool, cfg))
	r.Post("/api/auth/login", handlers.Login(pool, cfg))
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Protected routes — require valid JWT
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg.JWTSecret))

		r.Get("/api/me", func(w http.ResponseWriter, r *http.Request) {
			claims := middleware.GetClaims(r)
			if claims == nil {
				http.Error(w, "no claims in context", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"id":%d,"email":%q,"username":%q}`,
				claims.UserID, claims.Email, claims.Username)
		})

		// Master Data - Reference
		r.Get("/api/fsm/service-types", handlers.ListServiceTypes(pool))
		r.Post("/api/fsm/service-types", handlers.CreateServiceType(pool))
		r.Put("/api/fsm/service-types/{id}", handlers.UpdateServiceType(pool))
		r.Delete("/api/fsm/service-types/{id}", handlers.DeleteServiceType(pool))

		r.Get("/api/fsm/priority-levels", handlers.ListPriorityLevels(pool))
		r.Post("/api/fsm/priority-levels", handlers.CreatePriorityLevel(pool))
		r.Put("/api/fsm/priority-levels/{id}", handlers.UpdatePriorityLevel(pool))
		r.Delete("/api/fsm/priority-levels/{id}", handlers.DeletePriorityLevel(pool))

		r.Get("/api/fsm/skill-categories", handlers.ListSkillCategories(pool))
		r.Post("/api/fsm/skill-categories", handlers.CreateSkillCategory(pool))
		r.Put("/api/fsm/skill-categories/{id}", handlers.UpdateSkillCategory(pool))
		r.Delete("/api/fsm/skill-categories/{id}", handlers.DeleteSkillCategory(pool))

		r.Get("/api/fsm/asset-categories", handlers.ListAssetCategories(pool))
		r.Post("/api/fsm/asset-categories", handlers.CreateAssetCategory(pool))
		r.Put("/api/fsm/asset-categories/{id}", handlers.UpdateAssetCategory(pool))
		r.Delete("/api/fsm/asset-categories/{id}", handlers.DeleteAssetCategory(pool))

		// Customers
		r.Get("/api/fsm/customers", handlers.ListCustomers(pool))
		r.Get("/api/fsm/customers/{id}", handlers.GetCustomer(pool))
		r.Post("/api/fsm/customers", handlers.CreateCustomer(pool))
		r.Put("/api/fsm/customers/{id}", handlers.UpdateCustomer(pool))
		r.Delete("/api/fsm/customers/{id}", handlers.DeleteCustomer(pool))
		r.Get("/api/fsm/customers/{id}/sites", handlers.ListCustomerSitesByCustomer(pool))

		// Customer Sites
		r.Get("/api/fsm/customer-sites", handlers.ListCustomerSites(pool))
		r.Post("/api/fsm/customer-sites", handlers.CreateCustomerSite(pool))
		r.Put("/api/fsm/customer-sites/{id}", handlers.UpdateCustomerSite(pool))
		r.Delete("/api/fsm/customer-sites/{id}", handlers.DeleteCustomerSite(pool))

		// Users (for technician linking)
		r.Get("/api/fsm/users", handlers.ListUsers(pool))

		// Technicians
		r.Get("/api/fsm/technicians", handlers.ListTechnicians(pool))
		r.Get("/api/fsm/technicians/{id}", handlers.GetTechnician(pool))
		r.Post("/api/fsm/technicians", handlers.CreateTechnician(pool))
		r.Put("/api/fsm/technicians/{id}", handlers.UpdateTechnician(pool))
		r.Delete("/api/fsm/technicians/{id}", handlers.DeleteTechnician(pool))

		// Work Orders
		r.Get("/api/fsm/work-orders", handlers.ListWorkOrders(pool))
		r.Get("/api/fsm/work-orders/{id}", handlers.GetWorkOrder(pool))
		r.Post("/api/fsm/work-orders", handlers.CreateWorkOrder(pool))
		r.Put("/api/fsm/work-orders/{id}", handlers.UpdateWorkOrder(pool))
		r.Patch("/api/fsm/work-orders/{id}/status", handlers.UpdateWorkOrderStatus(pool))

		// Assignments
		r.Get("/api/fsm/work-orders/{id}/assignments", handlers.ListAssignments(pool))
		r.Post("/api/fsm/work-orders/{id}/assignments", handlers.AddAssignment(pool))
		r.Delete("/api/fsm/work-orders/{id}/assignments/{technicianId}", handlers.RemoveAssignment(pool))

		// Assets
		r.Get("/api/fsm/assets", handlers.ListAssets(pool))
		r.Get("/api/fsm/assets/{id}", handlers.GetAsset(pool))
		r.Post("/api/fsm/assets", handlers.CreateAsset(pool))
		r.Put("/api/fsm/assets/{id}", handlers.UpdateAsset(pool))
		r.Delete("/api/fsm/assets/{id}", handlers.DeleteAsset(pool))
		r.Get("/api/fsm/customer-sites/{id}/assets", handlers.ListAssetsBySite(pool))

		// Technician self-service
		r.Get("/api/fsm/my-work-orders", handlers.GetMyWorkOrders(pool))
	})

	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
