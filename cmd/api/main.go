package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/signalroot/signalroot/internal/auth"
	"github.com/signalroot/signalroot/internal/config"
	"github.com/signalroot/signalroot/internal/db"
	"github.com/signalroot/signalroot/internal/incident"
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

	authSvc := auth.NewService(cfg, database.Pool, logger)
	incidentRepo := db.NewIncidentRepo(database)
	signalRepo := db.NewSignalRepo(database)
	timelineRepo := db.NewTimelineRepo(database)
	jobRepo := db.NewJobRepo(database)

	api := &API{
		cfg:          cfg,
		logger:       logger,
		incidentRepo: incidentRepo,
		signalRepo:   signalRepo,
		timelineRepo: timelineRepo,
		jobRepo:      jobRepo,
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health checks (unauthenticated)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Health(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "not ready"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	// Auth routes
	r.Post("/auth/login", api.HandleLogin)
	r.Get("/auth/me", func(w http.ResponseWriter, r *http.Request) {
		authSvc.Middleware(http.HandlerFunc(api.HandleMe)).ServeHTTP(w, r)
	})

	// Authenticated API routes
	r.Group(func(r chi.Router) {
		r.Use(authSvc.Middleware)

		// Incidents
		r.Get("/api/v1/incidents", api.ListIncidents)
		r.Post("/api/v1/incidents", api.CreateIncident)
		r.Get("/api/v1/incidents/{id}", api.GetIncident)
		r.Put("/api/v1/incidents/{id}", api.UpdateIncident)

		// Incident status transitions
		r.Post("/api/v1/incidents/{id}/acknowledge", api.AcknowledgeIncident)
		r.Post("/api/v1/incidents/{id}/investigate", api.InvestigateIncident)
		r.Post("/api/v1/incidents/{id}/mitigate", api.MitigateIncident)
		r.Post("/api/v1/incidents/{id}/resolve", api.ResolveIncident)
		r.Post("/api/v1/incidents/{id}/close", api.CloseIncident)
		r.Post("/api/v1/incidents/{id}/reopen", api.ReopenIncident)

		// Timeline
		r.Get("/api/v1/incidents/{id}/timeline", api.GetTimeline)
		r.Post("/api/v1/incidents/{id}/timeline", api.AddTimelineEvent)

		// Similar incidents
		r.Get("/api/v1/incidents/{id}/similar", api.GetSimilarIncidents)

		// RCA
		r.Post("/api/v1/incidents/{id}/rca", api.TriggerRCA)

		// Signals
		r.Get("/api/v1/signals", api.ListSignals)

		// Analytics
		r.Get("/api/v1/analytics/overview", api.AnalyticsOverview)
	})

	srv := &http.Server{Addr: cfg.APIAddr, Handler: r}

	go func() {
		logger.Info("API server starting", zap.String("addr", cfg.APIAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down API server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}

// API holds the API server state and dependencies.
type API struct {
	cfg          *config.Config
	logger       *zap.Logger
	incidentRepo *db.IncidentRepo
	signalRepo   *db.SignalRepo
	timelineRepo *db.TimelineRepo
	jobRepo      *db.JobRepo
}

func (a *API) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (a *API) respondError(w http.ResponseWriter, status int, msg string) {
	a.respondJSON(w, status, map[string]string{"error": msg})
}

func (a *API) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Simplified login for dev — in production use OAuth2
	a.respondJSON(w, http.StatusOK, map[string]string{"message": "use OAuth2 flow"})
}

func (a *API) HandleMe(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		a.respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	a.respondJSON(w, http.StatusOK, user)
}

func (a *API) ListIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := auth.OrgIDFromContext(r.Context())
	q := r.URL.Query()

	filter := db.IncidentFilter{
		Limit: 25,
	}
	if v := q.Get("status"); v != "" {
		filter.Statuses = strings.Split(v, ",")
	}
	if v := q.Get("severity"); v != "" {
		filter.Severities = strings.Split(v, ",")
	}
	if v := q.Get("service"); v != "" {
		filter.Service = v
	}
	if v := q.Get("q"); v != "" {
		filter.SearchQuery = v
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &t
		}
	}

	incidents, total, err := a.incidentRepo.List(r.Context(), orgID, filter)
	if err != nil {
		a.logger.Error("failed to list incidents", zap.Error(err))
		a.respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if incidents == nil {
		incidents = []incident.Incident{}
	}

	a.respondJSON(w, http.StatusOK, incident.PaginatedResponse{
		Data: incidents,
		Meta: incident.PaginationMeta{
			HasMore: len(incidents) == filter.Limit,
			Total:   total,
		},
	})
}

func (a *API) CreateIncident(w http.ResponseWriter, r *http.Request) {
	orgID := auth.OrgIDFromContext(r.Context())
	user := auth.UserFromContext(r.Context())

	var req struct {
		Title    string   `json:"title"`
		Severity string   `json:"severity"`
		Services []string `json:"services"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Title == "" {
		a.respondError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Severity == "" {
		req.Severity = incident.SeverityUnknown
	}

	now := time.Now().UTC()
	inc := &incident.Incident{
		ID:               uuid.New(),
		OrgID:            orgID,
		Title:            req.Title,
		Severity:         req.Severity,
		Status:           incident.StatusOpen,
		ServicesAffected: req.Services,
		Environments:     []string{"production"},
		Tags:             []string{},
		DetectedAt:       now,
		CreatedAt:        now,
		UpdatedAt:        now,
		CreatedBy:        &user.ID,
		Metadata:         map[string]interface{}{"manual": true},
	}

	if err := a.incidentRepo.Create(r.Context(), inc); err != nil {
		a.logger.Error("failed to create incident", zap.Error(err))
		a.respondError(w, http.StatusInternalServerError, "failed to create incident")
		return
	}

	a.respondJSON(w, http.StatusCreated, inc)
}

func (a *API) GetIncident(w http.ResponseWriter, r *http.Request) {
	orgID := auth.OrgIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	inc, err := a.incidentRepo.GetByID(r.Context(), orgID, id)
	if err != nil {
		a.respondError(w, http.StatusNotFound, "incident not found")
		return
	}
	a.respondJSON(w, http.StatusOK, inc)
}

func (a *API) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	orgID := auth.OrgIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	inc, err := a.incidentRepo.GetByID(r.Context(), orgID, id)
	if err != nil {
		a.respondError(w, http.StatusNotFound, "incident not found")
		return
	}

	var req struct {
		Title    *string  `json:"title"`
		Severity *string  `json:"severity"`
		Tags     []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Title != nil {
		inc.Title = *req.Title
	}
	if req.Severity != nil {
		inc.Severity = *req.Severity
	}
	if req.Tags != nil {
		inc.Tags = req.Tags
	}

	if err := a.incidentRepo.Update(r.Context(), inc); err != nil {
		a.respondError(w, http.StatusInternalServerError, "failed to update")
		return
	}
	a.respondJSON(w, http.StatusOK, inc)
}

func (a *API) transitionIncident(w http.ResponseWriter, r *http.Request, newStatus string) {
	orgID := auth.OrgIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	inc, err := a.incidentRepo.GetByID(r.Context(), orgID, id)
	if err != nil {
		a.respondError(w, http.StatusNotFound, "incident not found")
		return
	}

	if !incident.IsValidTransition(inc.Status, newStatus) {
		a.respondError(w, http.StatusConflict, "invalid status transition from "+inc.Status+" to "+newStatus)
		return
	}

	now := time.Now().UTC()
	inc.Status = newStatus
	switch newStatus {
	case incident.StatusAcknowledged:
		inc.AcknowledgedAt = &now
	case incident.StatusMitigated:
		inc.MitigatedAt = &now
	case incident.StatusResolved:
		inc.ResolvedAt = &now
	case incident.StatusClosed:
		inc.ClosedAt = &now
	case incident.StatusOpen:
		// Reopening
		inc.ResolvedAt = nil
		inc.ClosedAt = nil
	}

	if err := a.incidentRepo.Update(r.Context(), inc); err != nil {
		a.respondError(w, http.StatusInternalServerError, "failed to update")
		return
	}

	// Add timeline event
	user := auth.UserFromContext(r.Context())
	evt := &incident.TimelineEvent{
		ID:          uuid.New(),
		IncidentID:  inc.ID,
		EventType:   "status_change",
		ActorType:   strPtr("user"),
		ActorID:     &user.ID,
		Description: "Status changed to " + newStatus,
		Metadata:    map[string]interface{}{},
		OccurredAt:  now,
		CreatedAt:   now,
	}
	a.timelineRepo.Insert(r.Context(), evt)

	// If resolved, trigger RCA generation
	if newStatus == incident.StatusResolved {
		a.enqueueJob(r.Context(), "generate_rca", map[string]interface{}{
			"incident_id": inc.ID.String(),
			"org_id":      inc.OrgID.String(),
		})
	}

	a.respondJSON(w, http.StatusOK, inc)
}

func (a *API) AcknowledgeIncident(w http.ResponseWriter, r *http.Request) {
	a.transitionIncident(w, r, incident.StatusAcknowledged)
}
func (a *API) InvestigateIncident(w http.ResponseWriter, r *http.Request) {
	a.transitionIncident(w, r, incident.StatusInvestigating)
}
func (a *API) MitigateIncident(w http.ResponseWriter, r *http.Request) {
	a.transitionIncident(w, r, incident.StatusMitigated)
}
func (a *API) ResolveIncident(w http.ResponseWriter, r *http.Request) {
	a.transitionIncident(w, r, incident.StatusResolved)
}
func (a *API) CloseIncident(w http.ResponseWriter, r *http.Request) {
	a.transitionIncident(w, r, incident.StatusClosed)
}
func (a *API) ReopenIncident(w http.ResponseWriter, r *http.Request) {
	a.transitionIncident(w, r, incident.StatusOpen)
}

func (a *API) GetTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	events, err := a.timelineRepo.ListByIncident(r.Context(), id)
	if err != nil {
		a.respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if events == nil {
		events = []incident.TimelineEvent{}
	}
	a.respondJSON(w, http.StatusOK, map[string]interface{}{"data": events})
}

func (a *API) AddTimelineEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	user := auth.UserFromContext(r.Context())

	var req struct {
		Description string `json:"description"`
		EventType   string `json:"event_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.EventType == "" {
		req.EventType = "comment"
	}

	now := time.Now().UTC()
	evt := &incident.TimelineEvent{
		ID:          uuid.New(),
		IncidentID:  id,
		EventType:   req.EventType,
		ActorType:   strPtr("user"),
		ActorID:     &user.ID,
		Description: req.Description,
		Metadata:    map[string]interface{}{},
		OccurredAt:  now,
		CreatedAt:   now,
	}
	if err := a.timelineRepo.Insert(r.Context(), evt); err != nil {
		a.respondError(w, http.StatusInternalServerError, "failed to add event")
		return
	}
	a.respondJSON(w, http.StatusCreated, evt)
}

func (a *API) GetSimilarIncidents(w http.ResponseWriter, r *http.Request) {
	// TODO: call ML service for similarity search
	a.respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    []interface{}{},
		"message": "similarity search requires ML service",
	})
}

func (a *API) TriggerRCA(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		a.respondError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	orgID := auth.OrgIDFromContext(r.Context())

	a.enqueueJob(r.Context(), "generate_rca", map[string]interface{}{
		"incident_id": id.String(),
		"org_id":      orgID.String(),
	})

	a.respondJSON(w, http.StatusAccepted, map[string]string{
		"status":  "queued",
		"message": "RCA generation has been queued",
	})
}

func (a *API) ListSignals(w http.ResponseWriter, r *http.Request) {
	a.respondJSON(w, http.StatusOK, incident.PaginatedResponse{
		Data: []interface{}{},
		Meta: incident.PaginationMeta{HasMore: false, Total: 0},
	})
}

func (a *API) AnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	a.respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_incidents":     0,
		"open_incidents":      0,
		"mttr_minutes":        0,
		"mtta_minutes":        0,
		"incidents_this_week": 0,
	})
}

func (a *API) enqueueJob(ctx context.Context, jobType string, payload map[string]interface{}) {
	now := time.Now().UTC()
	job := &incident.Job{
		ID:          uuid.New(),
		JobType:     jobType,
		Payload:     payload,
		MaxAttempts: 3,
		RunAt:       now,
		CreatedAt:   now,
	}
	if err := a.jobRepo.Enqueue(ctx, job); err != nil {
		a.logger.Error("failed to enqueue job", zap.String("type", jobType), zap.Error(err))
	}
}

func strPtr(s string) *string { return &s }
