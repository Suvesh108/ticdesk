package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticDesk/internal/auth"
	"ticDesk/internal/config"
	"ticDesk/internal/handlers"
	"ticDesk/internal/repository"
	"ticDesk/internal/router"
	"ticDesk/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Printf("Warning: Database ping failed: %v", err)
	} else {
		log.Println("Successfully connected to PostgreSQL database")
	}

	// Initialize Services
	storageService := services.NewLocalStorageService("web/static/uploads")

	// Initialize Session Manager
	sessionManager := auth.InitSessionManager(cfg.SessionSecret)

	// Initialize Repositories
	userRepo := repository.NewUserRepository(dbPool)
	ticketRepo := repository.NewTicketRepository(dbPool)
	commentRepo := repository.NewCommentRepository(dbPool)
	attachmentRepo := repository.NewAttachmentRepository(dbPool)

	// Initialize Handlers
	authHandler := handlers.NewAuthHandler(userRepo)
	dashboardHandler := handlers.NewDashboardHandler()
	ticketHandler := handlers.NewTicketHandler(ticketRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo, attachmentRepo, ticketRepo, storageService)
	attachmentHandler := handlers.NewAttachmentHandler(attachmentRepo, ticketRepo)

	// Build Router
	r := router.New(
		sessionManager,
		authHandler,
		dashboardHandler,
		ticketHandler,
		commentHandler,
		attachmentHandler,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("ticDesk server starting on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down ticDesk server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("ticDesk server stopped gracefully.")
}
