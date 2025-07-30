package prometheus

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fmotalleb/dor-radar/config"
)

type PromCollector struct {
	config *config.Collector
}

func New(collectors *config.Collector) *PromCollector {
	return &PromCollector{
		config: collectors,
	}
}

// ServeHTTP implements http.Handler.
func (p *PromCollector) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	windowStr := params.Get("window")
	if windowStr == "" {
		windowStr = "10"
	}
	window, err := strconv.Atoi(windowStr)
	if err != nil {
		http.Error(res, "Failed to parse request", http.StatusBadRequest)
	}

	minimum := false
	if params.Get("method") == "min" {
		minimum = true
	}
	data, err := GetData(req.Context(), *p.config, window, minimum)
	if err != nil {
		http.Error(res, "Failed to get data", http.StatusInternalServerError)
	}
	res.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(res).Encode(data); err != nil {
		http.Error(res, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
