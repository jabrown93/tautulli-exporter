// tautulli-exporter exposes Tautulli activity statistics as Prometheus
// metrics. It queries Tautulli's v2 API (cmd=get_activity) on every scrape.
//
// Configuration (environment):
//
//	TAUTULLI_URI      base URL of the Tautulli instance, e.g. http://tautulli:8181 (required)
//	TAUTULLI_API_KEY  Tautulli API key (required)
//	SERVE_PORT        port to listen on (default 9487)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "tautulli"

type activityResponse struct {
	Response struct {
		Result string `json:"result"`
		Data   struct {
			StreamCount             json.Number `json:"stream_count"`
			StreamCountDirectPlay   json.Number `json:"stream_count_direct_play"`
			StreamCountDirectStream json.Number `json:"stream_count_direct_stream"`
			StreamCountTranscode    json.Number `json:"stream_count_transcode"`
			TotalBandwidth          json.Number `json:"total_bandwidth"`
			LanBandwidth            json.Number `json:"lan_bandwidth"`
			WanBandwidth            json.Number `json:"wan_bandwidth"`
		} `json:"data"`
	} `json:"response"`
}

type collector struct {
	apiURL string
	client *http.Client

	up               *prometheus.Desc
	streamCount      *prometheus.Desc
	streamDirectPlay *prometheus.Desc
	streamDirectStrm *prometheus.Desc
	streamTranscode  *prometheus.Desc
	bandwidthTotal   *prometheus.Desc
	bandwidthLan     *prometheus.Desc
	bandwidthWan     *prometheus.Desc
}

func newCollector(apiURL string) *collector {
	desc := func(name, help string) *prometheus.Desc {
		return prometheus.NewDesc(namespace+"_"+name, help, nil, nil)
	}
	return &collector{
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},

		up:               desc("up", "Whether the last Tautulli API query succeeded (1) or failed (0)."),
		streamCount:      desc("stream_count", "Current number of active streams."),
		streamDirectPlay: desc("stream_direct_play", "Current number of direct-play streams."),
		streamDirectStrm: desc("stream_direct_stream", "Current number of direct streams."),
		streamTranscode:  desc("stream_count_transcode", "Current number of transcoding streams."),
		bandwidthTotal:   desc("bandwidth_total", "Total streaming bandwidth as reported by Tautulli (kbps)."),
		bandwidthLan:     desc("bandwidth_lan", "LAN streaming bandwidth as reported by Tautulli (kbps)."),
		bandwidthWan:     desc("bandwidth_wan", "WAN streaming bandwidth as reported by Tautulli (kbps)."),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.streamCount
	ch <- c.streamDirectPlay
	ch <- c.streamDirectStrm
	ch <- c.streamTranscode
	ch <- c.bandwidthTotal
	ch <- c.bandwidthLan
	ch <- c.bandwidthWan
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	gauge := func(d *prometheus.Desc, v float64) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v)
	}

	act, err := c.fetchActivity()
	if err != nil {
		log.Printf("tautulli API query failed: %v", err)
		gauge(c.up, 0)
		return
	}
	gauge(c.up, 1)

	num := func(n json.Number) float64 {
		f, err := n.Float64()
		if err != nil {
			return 0
		}
		return f
	}
	d := act.Response.Data
	gauge(c.streamCount, num(d.StreamCount))
	gauge(c.streamDirectPlay, num(d.StreamCountDirectPlay))
	gauge(c.streamDirectStrm, num(d.StreamCountDirectStream))
	gauge(c.streamTranscode, num(d.StreamCountTranscode))
	gauge(c.bandwidthTotal, num(d.TotalBandwidth))
	gauge(c.bandwidthLan, num(d.LanBandwidth))
	gauge(c.bandwidthWan, num(d.WanBandwidth))
}

func (c *collector) fetchActivity() (*activityResponse, error) {
	resp, err := c.client.Get(c.apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	var act activityResponse
	if err := json.NewDecoder(resp.Body).Decode(&act); err != nil {
		return nil, err
	}
	if act.Response.Result != "success" {
		return nil, fmt.Errorf("API result %q", act.Response.Result)
	}
	return &act, nil
}

func main() {
	uri := os.Getenv("TAUTULLI_URI")
	apiKey := os.Getenv("TAUTULLI_API_KEY")
	if uri == "" || apiKey == "" {
		log.Fatal("TAUTULLI_URI and TAUTULLI_API_KEY are required")
	}
	port := os.Getenv("SERVE_PORT")
	if port == "" {
		port = "9487"
	}

	apiURL := fmt.Sprintf("%s/api/v2?apikey=%s&cmd=get_activity", uri, url.QueryEscape(apiKey))

	reg := prometheus.NewRegistry()
	reg.MustRegister(newCollector(apiURL))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
