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
	url := baseUrl
	url = url.JoinPath("api", "v1", "query")
	query := url.Query()
	query.Add("query", "min_over_time(probe_success[10m])")
	url.RawQuery = query.Encode()
	// Prepare request URL with query param
	reqURL := url.String()
	cl := new(http.Client)
	cl.Timeout = time.Second * 20
	resp, err := cl.Get(reqURL)
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
		return nil, err
	}

	// Proceed to transform promResp.Data.Result
	nodesMap := make(map[string]int) // map node name to id
	var nodes []map[string]interface{}
	var connections []map[string]interface{}
	nodeID := 0

	for _, result := range promResp.Data.Result {
		// Example label usage: use hostname as node name
		hostname := result.Metric["hostname"]
		target := result.Metric["target"]
		valueStr := fmt.Sprintf("%v", result.Value[1])

		// Add hostname node if not exists
		if _, ok := nodesMap[hostname]; !ok {
			nodesMap[hostname] = nodeID
			nodes = append(nodes, map[string]interface{}{
				"id":   nodeID,
				"name": hostname,
			})
			nodeID++
		}
		// Add target node if not exists
		if _, ok := nodesMap[target]; !ok {
			nodesMap[target] = nodeID
			nodes = append(nodes, map[string]interface{}{
				"id":   nodeID,
				"name": target,
			})
			nodeID++
		}

		// Parse value (should be float, 0 or 1)
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

	if err != nil {
		fmt.Println("Error marshaling output JSON:", err)
		return nil, err
	}

	return output, nil
}
