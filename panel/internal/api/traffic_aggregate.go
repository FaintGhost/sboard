package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"sboard/panel/internal/db"
	"sboard/panel/internal/traffic"
)

type trafficNodeSummaryDTO struct {
	NodeID         int64  `json:"node_id"`
	Upload         int64  `json:"upload"`
	Download       int64  `json:"download"`
	LastRecordedAt string `json:"last_recorded_at"`
	Samples        int64  `json:"samples"`
	Inbounds       int64  `json:"inbounds"`
}

type trafficTotalSummaryDTO struct {
	Upload         int64  `json:"upload"`
	Download       int64  `json:"download"`
	LastRecordedAt string `json:"last_recorded_at"`
	Samples        int64  `json:"samples"`
	Nodes          int64  `json:"nodes"`
	Inbounds       int64  `json:"inbounds"`
}

type trafficTimeseriesPointDTO struct {
	BucketStart string `json:"bucket_start"`
	Upload      int64  `json:"upload"`
	Download    int64  `json:"download"`
}

func TrafficNodesSummary(store *db.Store) gin.HandlerFunc {
	p := traffic.NewSQLiteProvider(store)
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		window := parseWindowOrDefault(c.Query("window"), 24*time.Hour)
		if window < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window"})
			return
		}
		items, err := p.NodesSummary(c.Request.Context(), window)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out := make([]trafficNodeSummaryDTO, 0, len(items))
		for _, it := range items {
			out = append(out, trafficNodeSummaryDTO{
				NodeID:         it.NodeID,
				Upload:         it.Upload,
				Download:       it.Download,
				LastRecordedAt: timeRFC3339OrEmpty(it.LastRecordedAt),
				Samples:        it.Samples,
				Inbounds:       it.Inbounds,
			})
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func LegacyTrafficTotalSummary(store *db.Store) gin.HandlerFunc {
	p := traffic.NewSQLiteProvider(store)
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		window := parseWindowOrDefault(c.Query("window"), 24*time.Hour)
		if window < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window"})
			return
		}
		it, err := p.TotalSummary(c.Request.Context(), window)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": trafficTotalSummaryDTO{
			Upload:         it.Upload,
			Download:       it.Download,
			LastRecordedAt: timeRFC3339OrEmpty(it.LastRecordedAt),
			Samples:        it.Samples,
			Nodes:          it.Nodes,
			Inbounds:       it.Inbounds,
		}})
	}
}

func TrafficTimeseries(store *db.Store) gin.HandlerFunc {
	p := traffic.NewSQLiteProvider(store)
	return func(c *gin.Context) {
		if !ensureStore(c, store) {
			return
		}
		window := parseWindowOrDefault(c.Query("window"), 24*time.Hour)
		if window < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid window"})
			return
		}
		bucket := strings.TrimSpace(strings.ToLower(c.Query("bucket")))
		if bucket == "" {
			bucket = string(traffic.BucketHour)
		}
		var b traffic.Bucket
		switch bucket {
		case string(traffic.BucketMinute):
			b = traffic.BucketMinute
		case string(traffic.BucketHour):
			b = traffic.BucketHour
		case string(traffic.BucketDay):
			b = traffic.BucketDay
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket"})
			return
		}

		nodeID := int64(0)
		if v := strings.TrimSpace(c.Query("node_id")); v != "" {
			id, err := strconv.ParseInt(v, 10, 64)
			if err != nil || id < 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid node_id"})
				return
			}
			nodeID = id
		}

		items, err := p.Timeseries(c.Request.Context(), traffic.TimeseriesQuery{
			Window: window,
			Bucket: b,
			NodeID: nodeID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out := make([]trafficTimeseriesPointDTO, 0, len(items))
		for _, it := range items {
			out = append(out, trafficTimeseriesPointDTO{
				BucketStart: timeRFC3339OrEmpty(it.BucketStart),
				Upload:      it.Upload,
				Download:    it.Download,
			})
		}
		c.JSON(http.StatusOK, gin.H{"data": out})
	}
}

func parseWindowOrDefault(raw string, def time.Duration) time.Duration {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return def
	}

	if s == "all" {
		return 0
	}

	// Support "7d"/"30d" while keeping time.ParseDuration compatibility.
	if strings.HasSuffix(s, "d") {
		nStr := strings.TrimSuffix(s, "d")
		n, err := strconv.Atoi(nStr)
		if err != nil || n <= 0 {
			return -1
		}
		d := time.Duration(n) * 24 * time.Hour
		return clampWindow(d)
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return -1
	}
	return clampWindow(d)
}

func clampWindow(d time.Duration) time.Duration {
	// Protect backend from accidental huge scans.
	if d < time.Minute {
		return -1
	}
	if d > 90*24*time.Hour {
		return 90 * 24 * time.Hour
	}
	return d
}
