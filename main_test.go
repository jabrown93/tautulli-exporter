package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func TestCollectorFetchActivity(t *testing.T) {
	const apiKey = "test key+value"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2" {
			t.Errorf("path = %q, want /api/v2", r.URL.Path)
		}
		if got := r.URL.Query().Get("apikey"); got != apiKey {
			t.Errorf("apikey = %q, want %q", got, apiKey)
		}
		if got := r.URL.Query().Get("cmd"); got != "get_activity" {
			t.Errorf("cmd = %q, want get_activity", got)
		}
		_, _ = w.Write([]byte(readFixture(t, "activity_success.json")))
	}))
	defer server.Close()

	collector := newCollector(server.URL + "/api/v2?apikey=" + url.QueryEscape(apiKey) + "&cmd=get_activity")
	activity, err := collector.fetchActivity()
	if err != nil {
		t.Fatal(err)
	}
	if got := activity.Response.Data.StreamCount.String(); got != "3" {
		t.Errorf("stream count = %q, want 3", got)
	}
	if got := activity.Response.Data.TotalBandwidth.String(); got != "12000" {
		t.Errorf("total bandwidth = %q, want 12000", got)
	}
}

func TestCollectorCollectsActivityMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(readFixture(t, "activity_success.json")))
	}))
	defer server.Close()

	collector := newCollector(server.URL)
	if err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP tautulli_bandwidth_lan LAN streaming bandwidth as reported by Tautulli (kbps).
# TYPE tautulli_bandwidth_lan gauge
tautulli_bandwidth_lan 8000
# HELP tautulli_bandwidth_total Total streaming bandwidth as reported by Tautulli (kbps).
# TYPE tautulli_bandwidth_total gauge
tautulli_bandwidth_total 12000
# HELP tautulli_bandwidth_wan WAN streaming bandwidth as reported by Tautulli (kbps).
# TYPE tautulli_bandwidth_wan gauge
tautulli_bandwidth_wan 4000
# HELP tautulli_stream_count Current number of active streams.
# TYPE tautulli_stream_count gauge
tautulli_stream_count 3
# HELP tautulli_stream_count_transcode Current number of transcoding streams.
# TYPE tautulli_stream_count_transcode gauge
tautulli_stream_count_transcode 1
# HELP tautulli_stream_direct_play Current number of direct-play streams.
# TYPE tautulli_stream_direct_play gauge
tautulli_stream_direct_play 1
# HELP tautulli_stream_direct_stream Current number of direct streams.
# TYPE tautulli_stream_direct_stream gauge
tautulli_stream_direct_stream 1
# HELP tautulli_up Whether the last Tautulli API query succeeded (1) or failed (0).
# TYPE tautulli_up gauge
tautulli_up 1
`)); err != nil {
		t.Fatal(err)
	}
	if got := testutil.CollectAndCount(collector); got != 8 {
		t.Errorf("metric count = %d, want 8", got)
	}
}

func TestCollectorReportsDownOnAPIErrors(t *testing.T) {
	tests := []struct {
		name   string
		apiURL func(t *testing.T) string
	}{
		{
			name: "API down",
			apiURL: func(t *testing.T) string {
				t.Helper()
				server := httptest.NewServer(nil)
				url := server.URL
				server.Close()
				return url
			},
		},
		{
			name: "bad JSON",
			apiURL: func(t *testing.T) string {
				t.Helper()
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(readFixture(t, "activity_bad_json.txt")))
				}))
				t.Cleanup(server.Close)
				return server.URL
			},
		},
		{
			name: "bad API key",
			apiURL: func(t *testing.T) string {
				t.Helper()
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if got := r.URL.Query().Get("apikey"); got != "invalid" {
						t.Errorf("apikey = %q, want invalid", got)
					}
					_, _ = w.Write([]byte(readFixture(t, "activity_bad_api_key.json")))
				}))
				t.Cleanup(server.Close)
				return server.URL + "?apikey=invalid&cmd=get_activity"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collector := newCollector(test.apiURL(t))
			if err := testutil.CollectAndCompare(collector, strings.NewReader(`# HELP tautulli_up Whether the last Tautulli API query succeeded (1) or failed (0).
# TYPE tautulli_up gauge
tautulli_up 0
`)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
