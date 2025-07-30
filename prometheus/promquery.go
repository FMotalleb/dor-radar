package prometheus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/fmotalleb/dor-radar/config"
)

type PromResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  [2]interface{}    `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
type Json = map[string]interface{}

func GetData(ctx context.Context, cfg config.Collector, window int, minimum bool) (Json, error) {
	// Construct full query URL
	req, err := buildRequest(ctx, cfg, window, minimum)
	if err != nil {
		return nil, err
	}

	promResp, err := sendRequest(req)
	if err != nil {
		return nil, err
	}

	if promResp.Status != "success" {
		fmt.Println("Prometheus query failed")
		return nil, fmt.Errorf("query failed: %s", promResp.Status)
	}

	nodesMap := make(map[string]int)

	nodes, connections, err := extractFields(cfg, promResp, nodesMap)
	if err != nil {
		return nil, err
	}
	output := map[string]interface{}{
		"nodes":       nodes,
		"connections": connections,
	}

	return output, nil
}

func buildRequest(ctx context.Context, cfg config.Collector, window int, minimum bool) (*http.Request, error) {
	switch {
	case window < 1:
		return nil, errors.New("minimum range is 1 minute")
	case window > 60:
		return nil, errors.New("maximum range is 60 minutes")
	}
	method := "avg_over_time"
	if minimum {
		method = "min_over_time"
	}
	url := cfg.Target.JoinPath("api", "v1", "query")
	queryParams := url.Query()
	query := fmt.Sprintf("%s(probe_success%s[%dm])", method, cfg.Filter, window)
	queryParams.Add("query", query)
	url.RawQuery = queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Handle Basic Auth if credentials are provided
	if cfg.Target.User != nil {
		username := cfg.Target.User.Username()
		password, _ := cfg.Target.User.Password() // ok to ignore second return value
		req.SetBasicAuth(username, password)
	}
	return req, nil
}

func sendRequest(req *http.Request) (PromResponse, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error querying Prometheus:", err)
		return PromResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return PromResponse{}, err
	}

	var promResp PromResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return PromResponse{}, err
	}
	return promResp, nil
}

func extractFields(cfg config.Collector, promResp PromResponse, nodesMap map[string]int) ([]map[string]interface{}, []map[string]interface{}, error) {
	var nodes []map[string]interface{}
	var connections []map[string]interface{}
	nodeID := 0

	for _, result := range promResp.Data.Result {
		hostname := cfg.GetShape(result.Metric["hostname"])
		target := cfg.GetShape(result.Metric["target"])
		valueStr := fmt.Sprintf("%v", result.Value[1])

		if _, ok := nodesMap[hostname]; !ok {
			nodesMap[hostname] = nodeID
			nodes = append(nodes, map[string]interface{}{
				"id":    nodeID,
				"name":  hostname,
				"attrs": cfg.GetAttrs(result.Metric["hostname"]),
				"size":  cfg.GetSize(result.Metric["hostname"]),
			})
			nodeID++
		}
		if _, ok := nodesMap[target]; !ok {
			nodesMap[target] = nodeID
			nodes = append(nodes, map[string]interface{}{
				"id":    nodeID,
				"name":  target,
				"attrs": cfg.GetAttrs(result.Metric["target"]),
				"size":  cfg.GetSize(result.Metric["target"]),
			})
			nodeID++
		}

		var strength float64
		_, err := fmt.Sscanf(valueStr, "%f", &strength)
		if err != nil {
			return nil, nil, err
		}
		connections = append(connections, map[string]interface{}{
			"source":   nodesMap[hostname],
			"target":   nodesMap[target],
			"strength": strength,
		})
	}
	return nodes, updateStrength(connections), nil
}

func updateStrength(connections []Json) []Json {
	result := make([]Json, len(connections))
	copy(result, connections)
	for _, c := range result {
		c["strength"] = calculateWeakestLink(result, c["source"], c["strength"].(float64))
	}
	return result
}

func calculateWeakestLink(connections []Json, current interface{}, currentStrength float64) float64 {
	weakest := currentStrength

	for _, c := range connections {
		if c["target"] == current {
			if s, ok := c["strength"].(float64); ok {
				if s < weakest {
					weakest = s
				}
			}
		}
	}

	return weakest // fallback if no matching connection
}
