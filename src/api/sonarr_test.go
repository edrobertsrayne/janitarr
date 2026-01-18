package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSonarrClient_TestConnection(t *testing.T) {
	expected := SystemStatus{AppName: "Sonarr", Version: "4.0.0"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	result, err := client.TestConnection(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AppName != expected.AppName {
		t.Errorf("AppName = %q, want %q", result.AppName, expected.AppName)
	}
}

func TestSonarrClient_GetMissing(t *testing.T) {
	expected := PagedResponse[Episode]{
		Page:         1,
		PageSize:     50,
		TotalRecords: 2,
		Records: []Episode{
			{ID: 1, Title: "Pilot", SeriesTitle: "Show One", SeasonNumber: 1, EpisodeNumber: 1, Monitored: true},
			{ID: 2, Title: "Episode 2", SeriesTitle: "Show One", SeasonNumber: 1, EpisodeNumber: 2, Monitored: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
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
}

func TestSonarrClient_GetCutoffUnmet(t *testing.T) {
	expected := PagedResponse[Episode]{
		Page:         1,
		PageSize:     50,
		TotalRecords: 1,
		Records: []Episode{
			{ID: 3, Title: "Episode 3", SeriesTitle: "Show One", SeasonNumber: 1, EpisodeNumber: 3, Monitored: true, HasFile: true},
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

	client := NewSonarrClient(server.URL, "testapikey")
	result, err := client.GetCutoffUnmet(context.Background(), 1, 50)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d, want 1", result.TotalRecords)
	}
}

func TestSonarrClient_TriggerSearch(t *testing.T) {
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
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "EpisodeSearch", Status: "started"})
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	err := client.TriggerSearch(context.Background(), []int{1, 2, 3})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedBody["name"] != "EpisodeSearch" {
		t.Errorf("command name = %q, want EpisodeSearch", receivedBody["name"])
	}
	episodeIds, ok := receivedBody["episodeIds"].([]any)
	if !ok {
		t.Fatal("episodeIds not found in body")
	}
	if len(episodeIds) != 3 {
		t.Errorf("len(episodeIds) = %d, want 3", len(episodeIds))
	}
}

func TestSonarrClient_GetAllMissing_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedResponse[Episode]{
			Page:         1,
			PageSize:     100,
			TotalRecords: 2,
			Records: []Episode{
				{ID: 1, Title: "Pilot", SeriesTitle: "Show One", SeasonNumber: 1, EpisodeNumber: 1, Monitored: true},
				{ID: 2, Title: "Episode 2", SeriesTitle: "Show One", SeasonNumber: 1, EpisodeNumber: 2, Monitored: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	items, err := client.GetAllMissing(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2", len(items))
	}
	if items[0].Type != "episode" {
		t.Errorf("item type = %q, want episode", items[0].Type)
	}
}

func TestSonarrClient_GetAllMissing_FormatsTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedResponse[Episode]{
			Page:         1,
			PageSize:     100,
			TotalRecords: 1,
			Records: []Episode{
				{ID: 1, Title: "Pilot", SeriesTitle: "Breaking Bad", SeasonNumber: 1, EpisodeNumber: 1, Monitored: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	items, err := client.GetAllMissing(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Breaking Bad - S01E01 - Pilot"
	if items[0].Title != expected {
		t.Errorf("title = %q, want %q", items[0].Title, expected)
	}
}

func TestSonarrClient_GetAllMissing_UsesSeriesObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedResponse[Episode]{
			Page:         1,
			PageSize:     100,
			TotalRecords: 1,
			Records: []Episode{
				{ID: 1, Title: "Pilot", Series: &Series{Title: "The Wire"}, SeasonNumber: 1, EpisodeNumber: 1, Monitored: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	items, err := client.GetAllMissing(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "The Wire - S01E01 - Pilot"
	if items[0].Title != expected {
		t.Errorf("title = %q, want %q", items[0].Title, expected)
	}
}

func TestSonarrClient_GetAllMissing_MultiplePages(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		page := r.URL.Query().Get("page")

		var resp PagedResponse[Episode]
		if page == "1" {
			resp = PagedResponse[Episode]{
				Page:         1,
				PageSize:     100,
				TotalRecords: 150,
				Records:      make([]Episode, 100),
			}
			for i := 0; i < 100; i++ {
				resp.Records[i] = Episode{ID: i + 1, Title: "Episode", SeriesTitle: "Show", SeasonNumber: 1, EpisodeNumber: i + 1, Monitored: true}
			}
		} else {
			resp = PagedResponse[Episode]{
				Page:         2,
				PageSize:     100,
				TotalRecords: 150,
				Records:      make([]Episode, 50),
			}
			for i := 0; i < 50; i++ {
				resp.Records[i] = Episode{ID: i + 101, Title: "Episode", SeriesTitle: "Show", SeasonNumber: 2, EpisodeNumber: i + 1, Monitored: true}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
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

func TestSonarrClient_GetAllCutoffUnmet_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedResponse[Episode]{
			Page:         1,
			PageSize:     100,
			TotalRecords: 1,
			Records: []Episode{
				{ID: 1, Title: "Pilot", SeriesTitle: "Show One", SeasonNumber: 1, EpisodeNumber: 1, Monitored: true, HasFile: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	items, err := client.GetAllCutoffUnmet(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestSonarrClient_GetMissing_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewSonarrClient(server.URL, "testapikey")
	_, err := client.GetMissing(context.Background(), 1, 50)

	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
