package emailwebhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/gateway"
	"github.com/sipeed/picoclaw/pkg/health"
	"github.com/sipeed/picoclaw/pkg/logger"
)

func init() {
	gateway.RegisterHealthExtension(register)
}

func register(server *health.Server, cfg *config.Config, workspace string) {
	if !cfg.EmailWebhook.Enabled {
		return
	}
	h := &handler{
		cfg:      cfg,
		emailDir: filepath.Join(workspace, "emails"),
	}
	server.RegisterHandler("/webhook/email", h.serve)
	logger.InfoCF("email", "Email webhook registered at /webhook/email", nil)
}

type handler struct {
	cfg      *config.Config
	emailDir string
}

func (h *handler) serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.cfg.EmailWebhook.Secret != "" && r.Header.Get("X-Pico-Auth") != h.cfg.EmailWebhook.Secret {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var msg struct {
		From    string `json:"from"`
		Subject string `json:"subject"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	logger.InfoCF("email", "Received email", map[string]any{
		"from":    msg.From,
		"subject": msg.Subject,
	})

	if err := os.MkdirAll(h.emailDir, 0o755); err != nil {
		logger.ErrorCF("email", "Failed to create email directory", map[string]any{"error": err.Error()})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	date := time.Now().Format("2006-01-02")
	filePath := filepath.Join(h.emailDir, date+".jsonl")

	record := struct {
		ReceivedAt string `json:"received_at"`
		From       string `json:"from"`
		Subject    string `json:"subject"`
		Content    string `json:"content"`
	}{
		ReceivedAt: time.Now().UTC().Format(time.RFC3339),
		From:       msg.From,
		Subject:    msg.Subject,
		Content:    msg.Content,
	}

	line, err := json.Marshal(record)
	if err != nil {
		logger.ErrorCF("email", "Failed to marshal email record", map[string]any{"error": err.Error()})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		logger.ErrorCF("email", "Failed to open email file", map[string]any{"error": err.Error()})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s\n", line); err != nil {
		logger.ErrorCF("email", "Failed to write email record", map[string]any{"error": err.Error()})
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
