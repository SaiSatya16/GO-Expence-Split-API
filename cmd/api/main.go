// cmd/api/main.go

package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"expense-sharing-api/internal/config"
	"expense-sharing-api/internal/handlers"
	"expense-sharing-api/internal/middleware"
	"expense-sharing-api/internal/repository"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "EXPENSE-SHARING-API ", log.LstdFlags)

	// Initialize database
	dbConfig := config.NewDBConfig()
	db, err := dbConfig.Connect()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Initialize database schema
	if err := dbConfig.InitSchema(db); err != nil {
		logger.Fatal(err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	groupHandler := handlers.NewGroupHandler(groupRepo)
	expenseHandler := handlers.NewExpenseHandler(expenseRepo, groupRepo)

	// Initialize router
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(middleware.LoggingMiddleware(logger))

	// Health check route
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	// Public routes
	router.HandleFunc("/api/register", userHandler.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/login", userHandler.Login).Methods(http.MethodPost)

	// Protected routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// Group routes
	api.HandleFunc("/groups", groupHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/groups", groupHandler.GetUserGroups).Methods(http.MethodGet)
	api.HandleFunc("/groups/{id}", groupHandler.GetByID).Methods(http.MethodGet)

	// Expense routes
	api.HandleFunc("/expenses", expenseHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/groups/{id}/expenses", expenseHandler.GetGroupExpenses).Methods(http.MethodGet)
	api.HandleFunc("/groups/{id}/balance", expenseHandler.GetBalanceSheet).Methods(http.MethodGet)

	// Configure server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server
	go func() {
		logger.Printf("Starting server on port %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Server shutting down...")
	srv.Close()
}
