package urlworker

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"ada-love-ide/internal/config/worker"
)

type Runtime struct {
	Config   worker.WorkerConfig
	BaseURL  string
	URLPaths worker.URLPaths
}

func New(cfg worker.WorkerConfig) *Runtime {
	baseURL := strings.TrimRight(cfg.Command, "/")
	paths := worker.DecodeURLPaths(cfg.Arguments)
	return &Runtime{
		Config:   cfg,
		BaseURL:  baseURL,
		URLPaths: paths,
	}
}

type Request struct {
	Method       string
	URL          string
	Headers      map[string]string
	Body         string
	Timeout      time.Duration
	Stream       bool
	ModelsFormat string
}

type chatBodyVars struct {
	Message string
	Model   string
	Stream  string
}

func (r *Runtime) BuildChatRequest(message, model string) Request {
	body := r.renderChatBody(message, model)

	return Request{
		Method: "POST",
		URL:    fmt.Sprintf("%s%s", r.BaseURL, r.URLPaths.ChatPath),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body:    body,
		Timeout: 120 * time.Second,
		Stream:  r.URLPaths.Stream,
	}
}

func (r *Runtime) renderChatBody(message, model string) string {
	tpl := r.URLPaths.ChatBodyTemplate
	if tpl == "" {
		tpl = `{"model":"{{.Model}}","messages":[{"role":"user","content":"{{.Message}}"}],"stream":{{.Stream}}}`
	}

	tmpl, err := template.New("chat").Parse(tpl)
	if err != nil {
		return fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"%s"}],"stream":%t}`, model, message, r.URLPaths.Stream)
	}

	streamVal := "false"
	if r.URLPaths.Stream {
		streamVal = "true"
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, chatBodyVars{
		Message: message,
		Model:   model,
		Stream:  streamVal,
	})
	if err != nil {
		return fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"%s"}],"stream":%t}`, model, message, r.URLPaths.Stream)
	}

	return buf.String()
}

func (r *Runtime) BuildModelsRequest() Request {
	modelsURL := r.Config.ModelsCommand
	if modelsURL == "" {
		modelsURL = fmt.Sprintf("%s%s", r.BaseURL, r.URLPaths.ModelsPath)
	} else if !strings.HasPrefix(modelsURL, "http") {
		modelsURL = fmt.Sprintf("%s%s", r.BaseURL, modelsURL)
	}
	return Request{
		Method:       "GET",
		URL:          modelsURL,
		Headers:      map[string]string{"Accept": "application/json"},
		Timeout:      30 * time.Second,
		ModelsFormat: r.URLPaths.ModelsFormat,
	}
}
