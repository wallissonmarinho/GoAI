package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/wallissonmarinho/GoAI/internal/app"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	internalKey := strings.TrimSpace(os.Getenv("GOAI_ADMIN_API_KEY"))
	if internalKey == "" {
		internalKey = strings.TrimSpace(os.Getenv("GOAI_INTERNAL_API_KEY"))
	}
	if internalKey == "" {
		log.Error("GOAI_ADMIN_API_KEY is required (legacy alias: GOAI_INTERNAL_API_KEY)")
		os.Exit(1)
	}
	keysCSV := os.Getenv("GOAI_GEMINI_API_KEYS")

	cfg := app.Config{
		InternalAPIKey: internalKey,
		GeminiKeysCSV:  keysCSV,
		GeminiModel:    strings.TrimSpace(os.Getenv("GOAI_GEMINI_MODEL")),
		GeminiBaseURL:  strings.TrimSpace(os.Getenv("GOAI_GEMINI_BASE_URL")),
		KeyCooldown:    parseCooldown(),
		UserAgent: func() string {
			if v := strings.TrimSpace(os.Getenv("GOAI_USER_AGENT")); v != "" {
				return v
			}
			return "GoAI/1.0"
		}(),
	}
	if cfg.GeminiModel == "" {
		cfg.GeminiModel = "gemini-2.0-flash"
	}

	pool := geminiKeyCount(keysCSV)
	if pool == 0 {
		log.Error("GOAI_GEMINI_API_KEYS must list at least one non-empty key")
		os.Exit(1)
	}

	engine := app.Wire(cfg)

	addr := listenAddr()
	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("goai listening", slog.String("addr", addr), slog.String("model", cfg.GeminiModel), slog.Int("gemini_keys", pool))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Warn("shutdown", slog.Any("err", err))
	}
}

func geminiKeyCount(csv string) int {
	n := 0
	for _, s := range strings.Split(csv, ",") {
		if strings.TrimSpace(s) != "" {
			n++
		}
	}
	return n
}

func listenAddr() string {
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		return ":" + p
	}
	if a := strings.TrimSpace(os.Getenv("GOAI_HTTP_ADDR")); a != "" {
		return a
	}
	return ":8088"
}

func parseCooldown() time.Duration {
	s := strings.TrimSpace(os.Getenv("GOAI_GEMINI_KEY_COOLDOWN_SEC"))
	if s == "" {
		return 60 * time.Second
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 60 * time.Second
	}
	return time.Duration(n) * time.Second
}
