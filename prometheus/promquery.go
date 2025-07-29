package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
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

func GetData(ctx context.Context, baseUrl *url.URL) (map[string]interface{}, error) {
	// Construct full query URL
	url := baseUrl.JoinPath("api", "v1", "query")
	query := url.Query()
	query.Add("query", "min_over_time(probe_success[10m])")
	url.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	// Handle Basic Auth if credentials are provided
	if baseUrl.User != nil {
		username := baseUrl.User.Username()
		password, _ := baseUrl.User.Password() // ok to ignore second return value
		req.SetBasicAuth(username, password)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error querying Prometheus:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil, err
	}

	var promResp PromResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	if promResp.Status != "success" {
		fmt.Println("Prometheus query failed")
		return nil, fmt.Errorf("query failed: %s", promResp.Status)
	}

	nodesMap := make(map[string]int)
	var nodes []map[string]interface{}
	var connections []map[string]interface{}
	nodeID := 0

	for _, result := range promResp.Data.Result {
		hostname := result.Metric["hostname"]
		target := result.Metric["target"]
		valueStr := fmt.Sprintf("%v", result.Value[1])

		if _, ok := nodesMap[hostname]; !ok {
			nodesMap[hostname] = nodeID
			nodes = append(nodes, map[string]interface{}{
				"id":   nodeID,
				"name": hostname,
			})
			nodeID++
		}
		if _, ok := nodesMap[target]; !ok {
			nodesMap[target] = nodeID
			nodes = append(nodes, map[string]interface{}{
				"id":   nodeID,
				"name": target,
			})
			nodeID++
		}

		var strength float64
		fmt.Sscanf(valueStr, "%f", &strength)

		connections = append(connections, map[string]interface{}{
			"source":   nodesMap[hostname],
			"target":   nodesMap[target],
			"strength": strength,
		})
	}

	output := map[string]interface{}{
		"nodes":       nodes,
		"connections": connections,
	}

	return output, nil
}
