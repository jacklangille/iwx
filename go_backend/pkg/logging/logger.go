package logging

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"iwx/go_backend/internal/requestctx"
)

var serviceName string
var directWriter io.Writer
var directWriteMu sync.Mutex

func Setup(nextServiceName, commandDir string) (func() error, error) {
	if err := os.MkdirAll(commandDir, 0o755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(commandDir, nextServiceName+".log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	output := io.MultiWriter(os.Stdout, file)
	directWriter = output
	serviceName = nextServiceName
	log.SetFlags(0)
	log.SetOutput(&jsonWriter{
		output:  output,
		service: nextServiceName,
		writeMu: &sync.Mutex{},
	})
	Info(context.Background(), "logger_initialized", "file", logPath)

	return file.Close, nil
}

func Info(ctx context.Context, message string, kv ...any) {
	writeEntry("info", ctx, nil, message, kv...)
}

func Error(ctx context.Context, message string, err error, kv ...any) {
	writeEntry("error", ctx, err, message, kv...)
}

type jsonWriter struct {
	output  io.Writer
	service string
	writeMu *sync.Mutex
}

func (w *jsonWriter) Write(p []byte) (int, error) {
	raw := strings.TrimSpace(string(p))
	if raw == "" {
		return len(p), nil
	}

	entry := map[string]any{
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"level":   "info",
		"service": w.service,
	}

	message, fields := parseLogLine(raw)
	entry["msg"] = message
	for key, value := range fields {
		entry[key] = value
	}

	body, err := json.Marshal(entry)
	if err != nil {
		return 0, err
	}

	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	if _, err := w.output.Write(append(body, '\n')); err != nil {
		return 0, err
	}

	return len(p), nil
}

func writeEntry(level string, ctx context.Context, err error, message string, kv ...any) {
	entry := map[string]any{
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"level":   level,
		"service": serviceName,
		"msg":     strings.TrimSpace(message),
	}

	if requestID := requestctx.RequestID(ctx); requestID != "" {
		entry["request_id"] = requestID
	}
	if traceID := requestctx.TraceID(ctx); traceID != "" {
		entry["trace_id"] = traceID
	}
	if userID, ok := requestctx.UserID(ctx); ok && userID > 0 {
		entry["user_id"] = userID
	}
	if err != nil {
		entry["error"] = err.Error()
	}

	for index := 0; index+1 < len(kv); index += 2 {
		key, ok := kv[index].(string)
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		entry[key] = kv[index+1]
	}

	body, marshalErr := json.Marshal(entry)
	if marshalErr != nil {
		log.Printf("structured_log_marshal_failed error=%v", marshalErr)
		return
	}

	if directWriter == nil {
		directWriter = os.Stdout
	}
	directWriteMu.Lock()
	defer directWriteMu.Unlock()
	_, _ = directWriter.Write(append(body, '\n'))
}

func parseLogLine(raw string) (string, map[string]string) {
	tokens := strings.Fields(raw)
	if len(tokens) == 0 {
		return raw, map[string]string{"raw": raw}
	}

	messageTokens := make([]string, 0, len(tokens))
	fields := map[string]string{}
	foundField := false

	for _, token := range tokens {
		if key, value, ok := strings.Cut(token, "="); ok && strings.TrimSpace(key) != "" {
			foundField = true
			fields[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "\"")
			continue
		}

		if !foundField {
			messageTokens = append(messageTokens, token)
			continue
		}

		messageTokens = append(messageTokens, token)
	}

	message := strings.TrimSpace(strings.Join(messageTokens, " "))
	if message == "" {
		message = raw
	}
	if _, exists := fields["raw"]; !exists {
		fields["raw"] = raw
	}

	return message, fields
}
