package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

func TestVolcengineProvider_SubmitAndQuerySucceeded(t *testing.T) {
	var createBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/contents/generations/tasks":
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "volc-task-1"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/contents/generations/tasks/volc-task-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "volc-task-1",
				"status": "succeeded",
				"usage":  map[string]any{"completion_tokens": 49680},
				"content": map[string]any{
					"video_url": "https://example.com/out.mp4",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := NewVolcengineProvider(VolcengineConfig{APIKey: "test-key", BaseURL: srv.URL}, srv.Client())
	task := &media.Task{
		Model: "doubao-seedance-2.0",
		RequestParams: map[string]any{
			"prompt":     "a cat running",
			"duration":   5,
			"resolution": "720p",
			"ratio":      "16:9",
		},
	}

	id, err := p.Submit(context.Background(), task)
	if err != nil || id != "volc-task-1" {
		t.Fatalf("submit: id=%q err=%v", id, err)
	}
	if createBody["model"] != "doubao-seedance-2.0" {
		t.Fatalf("expected model in body, got %#v", createBody["model"])
	}
	content, ok := createBody["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("expected content from prompt, got %#v", createBody["content"])
	}

	task.UpstreamTaskID = id
	st, err := p.QueryStatus(context.Background(), task)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if st.State != media.ProviderSucceeded {
		t.Fatalf("expected succeeded, got %s", st.State)
	}
	if st.Usage.VideoTokens != 49680 {
		t.Fatalf("expected tokens 49680, got %d", st.Usage.VideoTokens)
	}
	if st.ResultURL != "https://example.com/out.mp4" {
		t.Fatalf("expected result url from content.video_url, got %q", st.ResultURL)
	}
}

func TestBuildVolcengineSubmitBodyMergesPromptIntoExistingContent(t *testing.T) {
	body := buildVolcengineSubmitBody(&media.Task{
		Model: "dreamina-seedance-2-0-260128",
		RequestParams: map[string]any{
			"prompt": "a kitten yawns",
			"content": []any{
				map[string]any{
					"type":      "image_url",
					"image_url": map[string]any{"url": "https://example.com/ref.jpg"},
				},
			},
		},
	})

	if _, ok := body["prompt"]; ok {
		t.Fatalf("prompt should be converted into content, got %#v", body["prompt"])
	}
	content, ok := body["content"].([]any)
	if !ok || len(content) != 2 {
		t.Fatalf("expected prompt plus reference content, got %#v", body["content"])
	}
	text, ok := content[0].(map[string]any)
	if !ok || text["type"] != "text" || text["text"] != "a kitten yawns" {
		t.Fatalf("expected text content first, got %#v", content[0])
	}
	ref, ok := content[1].(map[string]any)
	if !ok || ref["type"] != "image_url" {
		t.Fatalf("expected reference content second, got %#v", content[1])
	}
}

func TestResultURLFromResponse_Shapes(t *testing.T) {
	cases := []struct {
		name string
		raw  map[string]any
		want string
	}{
		{"content.video_url", map[string]any{"content": map[string]any{"video_url": "a"}}, "a"},
		{"top_video_url", map[string]any{"video_url": "b"}, "b"},
		{"output.video_url", map[string]any{"output": map[string]any{"video_url": "c"}}, "c"},
		{"data[0].url", map[string]any{"data": []any{map[string]any{"url": "d"}}}, "d"},
		{"missing", map[string]any{"status": "succeeded"}, ""},
		{"nil", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resultURLFromResponse(tc.raw); got != tc.want {
				t.Fatalf("resultURLFromResponse(%s)=%q want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestOpenRouterProvider_SubmitAndQueryCompleted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/videos":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "or-job-1", "status": "pending"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/videos/or-job-1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "or-job-1",
				"status": "completed",
				"usage":  map[string]any{"video_tokens": 12000},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := NewOpenRouterProvider(OpenRouterConfig{APIKey: "sk-or", BaseURL: srv.URL}, srv.Client())
	task := &media.Task{
		Model: "bytedance/seedance-2.0",
		RequestParams: map[string]any{
			"prompt":       "ocean waves",
			"duration":     8,
			"resolution":   "720p",
			"aspect_ratio": "16:9",
		},
	}
	id, err := p.Submit(context.Background(), task)
	if err != nil || id != "or-job-1" {
		t.Fatalf("submit: id=%q err=%v", id, err)
	}
	task.UpstreamTaskID = id
	st, err := p.QueryStatus(context.Background(), task)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if st.State != media.ProviderSucceeded || st.Usage.VideoTokens != 12000 {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestRegistry_RoutesByModel(t *testing.T) {
	srvVolc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "v1"})
	}))
	defer srvVolc.Close()
	srvOR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "o1"})
	}))
	defer srvOR.Close()

	reg := NewRegistry(Config{
		Volcengine: VolcengineConfig{APIKey: "vk", BaseURL: srvVolc.URL},
		OpenRouter: OpenRouterConfig{APIKey: "ok", BaseURL: srvOR.URL},
		HTTPClient: http.DefaultClient,
	})

	p1, err := reg.ProviderFor("doubao-seedance-2.0")
	if err != nil || p1 == nil {
		t.Fatalf("volcengine route: %v", err)
	}
	p2, err := reg.ProviderFor("bytedance/seedance-2.0")
	if err != nil || p2 == nil {
		t.Fatalf("openrouter route: %v", err)
	}
	if _, err := reg.ProviderFor("gpt-5.1"); err == nil {
		t.Fatal("expected unknown model error")
	}
}

func TestRegistry_SkipsUnconfiguredBackend(t *testing.T) {
	reg := NewRegistry(Config{})
	if _, err := reg.ProviderFor("doubao-seedance-2.0"); err == nil {
		t.Fatal("expected provider not found when volcengine key missing")
	}
}
