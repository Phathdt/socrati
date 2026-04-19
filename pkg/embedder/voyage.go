package embedder

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"socrati/pkg/logger"

	"github.com/imroc/req/v3"
)

const (
	defaultBaseURL    = "https://api.voyageai.com/v1"
	defaultModel      = "voyage-4-lite"
	defaultTimeout    = 5 * time.Second
	defaultMaxRetries = 3
	defaultMaxChars   = 8000
)

// VoyageConfig carries the knobs needed to talk to Voyage AI. Zero values fall
// back to sensible defaults so callers only need to supply the API key.
type VoyageConfig struct {
	APIKey     string
	Model      string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	MaxChars   int
	Client     *req.Client // optional; injected in tests
}

// VoyageEmbedder is a thin client around POST /v1/embeddings using req.Cool.
type VoyageEmbedder struct {
	cfg    VoyageConfig
	client *req.Client
	log    logger.Logger
}

// NewVoyage builds a VoyageEmbedder. Returns an error when the API key is
// missing — all other fields degrade to defaults.
func NewVoyage(cfg VoyageConfig, log logger.Logger) (*VoyageEmbedder, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("voyage: api key is required")
	}
	if cfg.Model == "" {
		cfg.Model = defaultModel
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	if cfg.MaxChars <= 0 {
		cfg.MaxChars = defaultMaxChars
	}

	client := cfg.Client
	if client == nil {
		client = req.C()
	}
	client.
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetCommonBearerAuthToken(cfg.APIKey).
		SetCommonContentType("application/json").
		SetCommonRetryCount(cfg.MaxRetries).
		SetCommonRetryBackoffInterval(200*time.Millisecond, 3*time.Second).
		AddCommonRetryCondition(func(resp *req.Response, err error) bool {
			if err != nil {
				return true
			}
			if resp == nil || resp.Response == nil {
				return true
			}
			code := resp.StatusCode
			return code == http.StatusTooManyRequests || code >= 500
		})

	if log != nil {
		client.AddCommonRetryHook(func(resp *req.Response, err error) {
			attempt := 0
			status := 0
			if resp != nil {
				attempt = resp.Request.RetryAttempt
				if resp.Response != nil {
					status = resp.StatusCode
				}
			}
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			log.Warn("voyage retry",
				"attempt", attempt,
				"status", status,
				"error", msg,
			)
		})
	}

	return &VoyageEmbedder{cfg: cfg, client: client, log: log}, nil
}

// Embed calls Voyage with no input_type hint.
func (v *VoyageEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return v.EmbedAs(ctx, text, InputTypeNone)
}

// EmbedAs lets callers pass an input_type hint (query/document).
func (v *VoyageEmbedder) EmbedAs(ctx context.Context, text string, kind InputType) ([]float32, error) {
	cleaned, err := v.prepareInput(text)
	if err != nil {
		return nil, err
	}

	body := voyageRequest{
		Input: []string{cleaned},
		Model: v.cfg.Model,
	}
	if kind != InputTypeNone {
		body.InputType = string(kind)
	}

	var out voyageResponse
	var errBody voyageError

	start := time.Now()
	resp, err := v.client.R().
		SetContext(ctx).
		SetBody(body).
		SetSuccessResult(&out).
		SetErrorResult(&errBody).
		Post("/embeddings")
	if err != nil {
		return nil, fmt.Errorf("voyage: request: %w", err)
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("voyage: status %d: %s", resp.StatusCode, errBody.String())
	}
	if len(out.Data) == 0 || len(out.Data[0].Embedding) == 0 {
		return nil, errors.New("voyage: empty embedding in response")
	}

	vec := out.Data[0].Embedding
	if v.log != nil {
		v.log.Info("voyage embed ok",
			"model", v.cfg.Model,
			"input_chars", len(cleaned),
			"vector_dim", len(vec),
			"tokens", out.Usage.TotalTokens,
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
	return vec, nil
}

func (v *VoyageEmbedder) prepareInput(text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", errors.New("voyage: input text must not be empty")
	}
	if len(trimmed) > v.cfg.MaxChars {
		if v.log != nil {
			v.log.Warn("voyage input truncated",
				"original_chars", len(trimmed),
				"max_chars", v.cfg.MaxChars,
			)
		}
		trimmed = trimmed[:v.cfg.MaxChars]
	}
	return trimmed, nil
}

type voyageRequest struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type,omitempty"`
}

type voyageResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type voyageError struct {
	Detail string `json:"detail"`
	Error  string `json:"error"`
}

func (e voyageError) String() string {
	if e.Error != "" {
		return e.Error
	}
	return e.Detail
}
