package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRadarrClient_TestConnection(t *testing.T) {
	expected := SystemStatus{AppName: "Radarr", Version: "5.0.0"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	result, err := client.TestConnection(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AppName != expected.AppName {
		t.Errorf("AppName = %q, want %q", result.AppName, expected.AppName)
	}
}

func TestRadarrClient_GetMissing(t *testing.T) {
	expected := PagedResponse[Movie]{
		Page:         1,
		PageSize:     50,
		TotalRecords: 2,
		Records: []Movie{
			{ID: 1, Title: "Movie One", Monitored: true},
			{ID: 2, Title: "Movie Two", Monitored: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("expected page=1, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("pageSize") != "50" {
			t.Errorf("expected pageSize=50, got %s", r.URL.Query().Get("pageSize"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	result, err := client.GetMissing(context.Background(), 1, 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalRecords != 2 {
		t.Errorf("TotalRecords = %d, want 2", result.TotalRecords)
	}
	if len(result.Records) != 2 {
		t.Errorf("len(Records) = %d, want 2", len(result.Records))
	}
	if result.Records[0].Title != "Movie One" {
		t.Errorf("first movie title = %q, want %q", result.Records[0].Title, "Movie One")
	}
}

func TestRadarrClient_GetCutoffUnmet(t *testing.T) {
	expected := PagedResponse[Movie]{
		Page:         1,
		PageSize:     50,
		TotalRecords: 1,
		Records: []Movie{
			{ID: 3, Title: "Movie Three", Monitored: true, HasFile: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/cutoff" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	result, err := client.GetCutoffUnmet(context.Background(), 1, 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", result.TotalRecords)
	}
}

func TestRadarrClient_TriggerSearch(t *testing.T) {
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/command" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "MoviesSearch", Status: "started"})
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	err := client.TriggerSearch(context.Background(), []int{1, 2, 3})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedBody["name"] != "MoviesSearch" {
		t.Errorf("command name = %q, want MoviesSearch", receivedBody["name"])
	}
	movieIds, ok := receivedBody["movieIds"].([]any)
	if !ok {
		t.Fatal("movieIds not found in body")
	}
	if len(movieIds) != 3 {
		t.Errorf("len(movieIds) = %d, want 3", len(movieIds))
	}
}

func TestRadarrClient_GetAllMissing_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedResponse[Movie]{
			Page:         1,
			PageSize:     100,
			TotalRecords: 2,
			Records: []Movie{
				{ID: 1, Title: "Movie One", Monitored: true},
				{ID: 2, Title: "Movie Two", Monitored: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	items, err := client.GetAllMissing(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
	if items[0].Type != "movie" {
		t.Errorf("item type = %q, want movie", items[0].Type)
	}
}

func TestRadarrClient_GetAllMissing_MultiplePages(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		page := r.URL.Query().Get("page")

		var resp PagedResponse[Movie]
		if page == "1" {
			resp = PagedResponse[Movie]{
				Page:         1,
				PageSize:     100,
				TotalRecords: 150,
				Records:      make([]Movie, 100),
			}
			for i := 0; i < 100; i++ {
				resp.Records[i] = Movie{ID: i + 1, Title: "Movie", Monitored: true}
			}
		} else {
			resp = PagedResponse[Movie]{
				Page:         2,
				PageSize:     100,
				TotalRecords: 150,
				Records:      make([]Movie, 50),
			}
			for i := 0; i < 50; i++ {
				resp.Records[i] = Movie{ID: i + 101, Title: "Movie", Monitored: true}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	items, err := client.GetAllMissing(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 150 {
		t.Errorf("len(items) = %d, want 150", len(items))
	}
	if requestCount != 2 {
		t.Errorf("requestCount = %d, want 2", requestCount)
	}
}

func TestRadarrClient_GetAllCutoffUnmet_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedResponse[Movie]{
			Page:         1,
			PageSize:     100,
			TotalRecords: 1,
			Records: []Movie{
				{ID: 1, Title: "Movie One", Monitored: true, HasFile: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	items, err := client.GetAllCutoffUnmet(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestRadarrClient_GetMissing_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewRadarrClient(server.URL, "testapikey")
	_, err := client.GetMissing(context.Background(), 1, 50)

	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
