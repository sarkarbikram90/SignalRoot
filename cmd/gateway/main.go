package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/signalroot/signalroot/internal/config"
	"github.com/signalroot/signalroot/internal/db"
	"github.com/signalroot/signalroot/internal/ingestion"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()
	ctx := context.Background()

	database, err := db.New(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	signalRepo := db.NewSignalRepo(database)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-PagerDuty-Signature", "X-Slack-Signature", "X-Slack-Request-Timestamp"},
		MaxAge:         300,
	}))

	gw := &Gateway{
		cfg:        cfg,
		signalRepo: signalRepo,
		logger:     logger,
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Health(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	r.Post("/webhooks/{integrationID}", gw.HandleWebhook)

	srv := &http.Server{
		Addr:    cfg.GatewayAddr,
		Handler: r,
	}

	go func() {
		logger.Info("gateway starting", zap.String("addr", cfg.GatewayAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gateway")
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

// Gateway handles inbound webhooks.
type Gateway struct {
	cfg        *config.Config
	signalRepo *db.SignalRepo
	logger     *zap.Logger
}

func (g *Gateway) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	integrationID := chi.URLParam(r, "integrationID")
	if integrationID == "" {
		http.Error(w, `{"error":"missing integration ID"}`, http.StatusBadRequest)
		return
	}

	intID, err := uuid.Parse(integrationID)
	if err != nil {
		http.Error(w, `{"error":"invalid integration ID"}`, http.StatusBadRequest)
		return
	}

	// Enforce payload size limit (1MB)
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			http.Error(w, `{"error":"payload too large"}`, http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}

	// Determine source type from integration config
	// For MVP, detect from headers
	sourceType := detectSourceType(r)
	adapter := ingestion.GetAdapter(sourceType)

	// Normalize payload
	canonicalSignals, err := adapter.Normalize(body)
	if err != nil {
		g.logger.Error("normalization failed",
			zap.String("source", sourceType),
			zap.Error(err),
		)
		http.Error(w, fmt.Sprintf(`{"error":"normalization failed: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// TODO: Determine org_id from integration lookup
	// For MVP, use a placeholder org ID
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	created := 0
	for _, cs := range canonicalSignals {
		sig := cs.ToSignal(orgID, &intID)
		if err := g.signalRepo.Insert(r.Context(), sig); err != nil {
			g.logger.Error("failed to insert signal",
				zap.String("source", sourceType),
				zap.Error(err),
			)
			continue
		}
		created++
	}

	g.logger.Info("webhook processed",
		zap.String("integration_id", integrationID),
		zap.String("source", sourceType),
		zap.Int("signals", created),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"signals": created,
	})
}

func detectSourceType(r *http.Request) string {
	if r.Header.Get("X-PagerDuty-Signature") != "" {
		return "pagerduty"
	}
	if r.Header.Get("X-Slack-Signature") != "" {
		return "slack"
	}
	return "webhook"
}
