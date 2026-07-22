package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JCheyoon/blog-backend/internal/auth"
	"github.com/JCheyoon/blog-backend/internal/category"
	"github.com/JCheyoon/blog-backend/internal/chat"
	"github.com/JCheyoon/blog-backend/internal/docs"
	"github.com/JCheyoon/blog-backend/internal/platform"
	"github.com/JCheyoon/blog-backend/internal/post"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := platform.LoadConfig()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := platform.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Wiring: handler -> service -> repository. Each domain owns its own
	// three files; main.go only assembles them. The mcp-server binary
	// (added later) will import post.Service directly and reuse this
	// exact business logic instead of duplicating it.
	chatClient := chat.NewClient(cfg.AnthropicAPIKey)
	// 20 questions per IP per hour on the /ask route - generous for a real
	// reader, cheap insurance against someone scripting your API key away.
	askLimiter := chat.NewRateLimiter(20, time.Hour)

	postRepo := post.NewRepository(db)
	postSvc := post.NewService(postRepo)
	postHandler := post.NewHandler(postSvc, chatClient)

	catRepo := category.NewRepository(db)
	catSvc := category.NewService(catRepo)
	catHandler := category.NewHandler(catSvc)

	authHandler := auth.NewHandler(cfg.AdminEmail, cfg.AdminPasswordHash, cfg.JWTSecret)
	requireAuth := auth.Middleware(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	authHandler.Register(mux)
	postHandler.Register(mux, requireAuth, askLimiter.Middleware)
	catHandler.Register(mux, requireAuth)
	docs.Register(mux)

	handler := withCORS(withLogging(mux))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // tighten to your frontend origin in production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
