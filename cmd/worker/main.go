package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/signalroot/signalroot/internal/config"
	"github.com/signalroot/signalroot/internal/db"
	"github.com/signalroot/signalroot/internal/incident"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.New(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer database.Close()

	jobRepo := db.NewJobRepo(database)
	incidentRepo := db.NewIncidentRepo(database)
	timelineRepo := db.NewTimelineRepo(database)

	w := &Worker{
		cfg:          cfg,
		logger:       logger,
		jobRepo:      jobRepo,
		incidentRepo: incidentRepo,
		timelineRepo: timelineRepo,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("worker starting")

	go w.Run(ctx)

	<-quit
	logger.Info("shutting down worker")
	cancel()
}

// Worker processes background jobs.
type Worker struct {
	cfg          *config.Config
	logger       *zap.Logger
	jobRepo      *db.JobRepo
	incidentRepo *db.IncidentRepo
	timelineRepo *db.TimelineRepo
}

func (w *Worker) Run(ctx context.Context) {
	// Retry failed jobs periodically
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, err := w.jobRepo.RetryFailed(ctx)
				if err != nil {
					w.logger.Error("failed to retry jobs", zap.Error(err))
				} else if count > 0 {
					w.logger.Info("retried failed jobs", zap.Int("count", count))
				}
			}
		}
	}()

	// Main job processing loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job, err := w.jobRepo.Dequeue(ctx)
			if err != nil {
				if err == pgx.ErrNoRows {
					time.Sleep(1 * time.Second)
					continue
				}
				w.logger.Error("failed to dequeue job", zap.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			w.logger.Info("processing job",
				zap.String("id", job.ID.String()),
				zap.String("type", job.JobType),
				zap.Int("attempt", job.Attempts),
			)

			if err := w.processJob(ctx, job); err != nil {
				w.logger.Error("job failed",
					zap.String("id", job.ID.String()),
					zap.String("type", job.JobType),
					zap.Error(err),
				)
				w.jobRepo.Fail(ctx, job.ID, err.Error())
			} else {
				w.jobRepo.Complete(ctx, job.ID)
				w.logger.Info("job completed",
					zap.String("id", job.ID.String()),
					zap.String("type", job.JobType),
				)
			}
		}
	}
}

func (w *Worker) processJob(ctx context.Context, job *incident.Job) error {
	switch job.JobType {
	case "generate_incident_summary":
		return w.generateSummary(ctx, job)
	case "generate_rca":
		return w.generateRCA(ctx, job)
	case "compute_incident_dna":
		return w.computeDNA(ctx, job)
	case "find_similar_incidents":
		return w.findSimilar(ctx, job)
	case "notify_slack":
		return w.notifySlack(ctx, job)
	default:
		w.logger.Warn("unknown job type", zap.String("type", job.JobType))
		return nil
	}
}

func (w *Worker) generateSummary(ctx context.Context, job *incident.Job) error {
	// TODO: Call Anthropic Claude API for summary generation
	w.logger.Info("generate_incident_summary: stub implementation",
		zap.Any("payload", job.Payload))
	return nil
}

func (w *Worker) generateRCA(ctx context.Context, job *incident.Job) error {
	// TODO: Call Anthropic Claude API for RCA
	w.logger.Info("generate_rca: stub implementation",
		zap.Any("payload", job.Payload))
	return nil
}

func (w *Worker) computeDNA(ctx context.Context, job *incident.Job) error {
	// TODO: Call ML service for embedding generation
	w.logger.Info("compute_incident_dna: stub implementation",
		zap.Any("payload", job.Payload))
	return nil
}

func (w *Worker) findSimilar(ctx context.Context, job *incident.Job) error {
	// TODO: Call ML service for similarity search
	w.logger.Info("find_similar_incidents: stub implementation",
		zap.Any("payload", job.Payload))
	return nil
}

func (w *Worker) notifySlack(ctx context.Context, job *incident.Job) error {
	// TODO: Send Slack notification
	w.logger.Info("notify_slack: stub implementation",
		zap.Any("payload", job.Payload))
	return nil
}
