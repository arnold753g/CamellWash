package httpapi

import (
	"log/slog"
	"net/http"

	"starter-backend/internal/config"
	"starter-backend/internal/handlers"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	services "starter-backend/internal/services"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	router := mux.NewRouter()
	authenticator := middleware.NewAuthenticator(db, cfg)
	authHandler, err := handlers.NewAuthHandler(db, cfg)
	if err != nil {
		return nil, err
	}
	profileHandler := handlers.NewProfileHandler(db)
	adminHandler := handlers.NewAdminHandler(db, cfg)
	catalogHandler := handlers.NewCatalogHandler(db)
	pricingHandler := handlers.NewPricingHandler(services.NewPricingService(db))
	customerHandler := handlers.NewCustomerHandler(services.NewCustomerService(db))
	orderHandler := handlers.NewOrderHandler(services.NewOrderService(db))
	paymentHandler := handlers.NewPaymentHandler(services.NewPaymentService(db))
	reportHandler := handlers.NewReportHandler(services.NewReportService(db))

	router.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.NewRateLimiter(cfg.APIRatePerMinute, cfg.TrustProxy).Middleware)

	publicAuth := api.PathPrefix("/auth").Subrouter()
	publicAuth.Use(middleware.NewRateLimiter(cfg.AuthRatePerMinute, cfg.TrustProxy).Middleware)
	publicAuth.Use(authenticator.OriginGuard)
	publicAuth.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	protected := api.NewRoute().Subrouter()
	protected.Use(authenticator.Middleware)
	protected.HandleFunc("/auth/session", authHandler.Session).Methods(http.MethodGet)
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods(http.MethodPost)
	protected.HandleFunc("/me", profileHandler.Get).Methods(http.MethodGet)
	protected.HandleFunc("/me", profileHandler.Update).Methods(http.MethodPatch)
	protected.HandleFunc("/me/password", authHandler.ChangePassword).Methods(http.MethodPost)

	catalogs := protected.PathPrefix("/catalogs").Subrouter()
	catalogs.HandleFunc("/vehicle-types", catalogHandler.ActiveVehicleTypes).Methods(http.MethodGet)
	catalogs.HandleFunc("/service-categories", catalogHandler.ActiveServiceCategories).Methods(http.MethodGet)
	catalogs.HandleFunc("/services", catalogHandler.ActiveServices).Methods(http.MethodGet)

	pricing := protected.PathPrefix("/pricing").Subrouter()
	pricing.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleManager))
	pricing.HandleFunc("/current", pricingHandler.Current).Methods(http.MethodGet)

	customers := protected.PathPrefix("/clients").Subrouter()
	customers.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleManager))
	customers.HandleFunc("", customerHandler.ListClients).Methods(http.MethodGet)
	customers.HandleFunc("", customerHandler.CreateClient).Methods(http.MethodPost)
	customers.HandleFunc("/{id:[0-9]+}", customerHandler.GetClient).Methods(http.MethodGet)
	customers.HandleFunc("/{id:[0-9]+}", customerHandler.UpdateClient).Methods(http.MethodPatch)
	customers.HandleFunc("/{id:[0-9]+}/vehicles", customerHandler.ListVehicles).Methods(http.MethodGet)
	customers.HandleFunc("/{id:[0-9]+}/vehicles", customerHandler.CreateVehicle).Methods(http.MethodPost)
	customers.HandleFunc("/{id:[0-9]+}/orders", orderHandler.ClientHistory).Methods(http.MethodGet)

	vehicles := protected.PathPrefix("/vehicles").Subrouter()
	vehicles.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleManager))
	vehicles.HandleFunc("/{id:[0-9]+}", customerHandler.GetVehicle).Methods(http.MethodGet)
	vehicles.HandleFunc("/{id:[0-9]+}", customerHandler.UpdateVehicle).Methods(http.MethodPatch)
	vehicles.HandleFunc("/{id:[0-9]+}/orders", orderHandler.VehicleHistory).Methods(http.MethodGet)

	orders := protected.PathPrefix("/orders").Subrouter()
	orders.HandleFunc("", orderHandler.List).Methods(http.MethodGet)
	orders.HandleFunc("", orderHandler.Create).Methods(http.MethodPost)
	orders.HandleFunc("/{id:[0-9]+}", orderHandler.Get).Methods(http.MethodGet)
	orders.HandleFunc("/{id:[0-9]+}", orderHandler.Update).Methods(http.MethodPatch)
	orders.HandleFunc("/{id:[0-9]+}/transitions", orderHandler.Transition).Methods(http.MethodPost)
	orders.HandleFunc("/{id:[0-9]+}/cancel", orderHandler.Cancel).Methods(http.MethodPost)
	orders.HandleFunc("/{id:[0-9]+}/payments", paymentHandler.ListForOrder).Methods(http.MethodGet)
	orders.HandleFunc("/{id:[0-9]+}/payments", paymentHandler.Register).Methods(http.MethodPost)

	payments := protected.PathPrefix("/payments").Subrouter()
	payments.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleManager))
	payments.HandleFunc("/{id:[0-9]+}/void", paymentHandler.Void).Methods(http.MethodPost)
	payments.HandleFunc("/{id:[0-9]+}/correct", paymentHandler.Correct).Methods(http.MethodPost)

	operations := protected.PathPrefix("/operations").Subrouter()
	operations.HandleFunc("/wash-staff", orderHandler.AvailableStaff).Methods(http.MethodGet)
	operations.HandleFunc("/service-options", orderHandler.AvailableServices).Methods(http.MethodGet)

	reports := protected.PathPrefix("/reports").Subrouter()
	reports.Use(middleware.RequireRoles(models.RoleAdmin, models.RoleManager))
	reports.HandleFunc("/operations", reportHandler.Operational).Methods(http.MethodGet)

	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireRoles(models.RoleAdmin))
	admin.HandleFunc("/dashboard", reportHandler.Dashboard).Methods(http.MethodGet)
	admin.HandleFunc("/reports/financial", reportHandler.Financial).Methods(http.MethodGet)
	admin.HandleFunc("/users", adminHandler.CreateUser).Methods(http.MethodPost)
	admin.HandleFunc("/users", adminHandler.ListUsers).Methods(http.MethodGet)
	admin.HandleFunc("/users/{id:[0-9]+}", adminHandler.GetUser).Methods(http.MethodGet)
	admin.HandleFunc("/users/{id:[0-9]+}", adminHandler.UpdateUser).Methods(http.MethodPatch)
	admin.HandleFunc("/users/{id:[0-9]+}/role", adminHandler.UpdateRole).Methods(http.MethodPatch)
	admin.HandleFunc("/users/{id:[0-9]+}/status", adminHandler.UpdateStatus).Methods(http.MethodPatch)
	admin.HandleFunc("/vehicle-types", catalogHandler.CreateVehicleType).Methods(http.MethodPost)
	admin.HandleFunc("/vehicle-types", catalogHandler.ListVehicleTypes).Methods(http.MethodGet)
	admin.HandleFunc("/vehicle-types/{id:[0-9]+}", catalogHandler.GetVehicleType).Methods(http.MethodGet)
	admin.HandleFunc("/vehicle-types/{id:[0-9]+}", catalogHandler.UpdateVehicleType).Methods(http.MethodPatch)
	admin.HandleFunc("/vehicle-types/{id:[0-9]+}/status", catalogHandler.UpdateVehicleTypeStatus).Methods(http.MethodPatch)
	admin.HandleFunc("/service-categories", catalogHandler.CreateServiceCategory).Methods(http.MethodPost)
	admin.HandleFunc("/service-categories", catalogHandler.ListServiceCategories).Methods(http.MethodGet)
	admin.HandleFunc("/service-categories/{id:[0-9]+}", catalogHandler.GetServiceCategory).Methods(http.MethodGet)
	admin.HandleFunc("/service-categories/{id:[0-9]+}", catalogHandler.UpdateServiceCategory).Methods(http.MethodPatch)
	admin.HandleFunc("/service-categories/{id:[0-9]+}/status", catalogHandler.UpdateServiceCategoryStatus).Methods(http.MethodPatch)
	admin.HandleFunc("/services", catalogHandler.CreateService).Methods(http.MethodPost)
	admin.HandleFunc("/services", catalogHandler.ListServices).Methods(http.MethodGet)
	admin.HandleFunc("/services/{id:[0-9]+}", catalogHandler.GetService).Methods(http.MethodGet)
	admin.HandleFunc("/services/{id:[0-9]+}", catalogHandler.UpdateService).Methods(http.MethodPatch)
	admin.HandleFunc("/services/{id:[0-9]+}/status", catalogHandler.UpdateServiceStatus).Methods(http.MethodPatch)
	admin.HandleFunc("/service-prices/matrix", pricingHandler.Matrix).Methods(http.MethodGet)
	admin.HandleFunc("/service-prices", pricingHandler.Create).Methods(http.MethodPost)
	admin.HandleFunc("/service-prices/{service_id:[0-9]+}/{vehicle_type_id:[0-9]+}", pricingHandler.Update).Methods(http.MethodPatch)
	admin.HandleFunc("/payment-reviews", paymentHandler.PendingReviews).Methods(http.MethodGet)
	admin.HandleFunc("/payment-reviews/{id:[0-9]+}", paymentHandler.Review).Methods(http.MethodPost)

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "Route was not found", nil)
	})
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respond.Error(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "HTTP method is not allowed", nil)
	})

	var handler http.Handler = router
	handler = middleware.CORS(cfg.AllowedOrigins).Handler(handler)
	handler = middleware.AccessLog(logger, cfg.TrustProxy)(handler)
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.Recover(logger)(handler)
	handler = middleware.RequestID(handler)
	return handler, nil
}
