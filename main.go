package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Data structure to hold JSON response
type WPStatus struct {
	Status                string  `json:"status"`
	WPStatusCode          int     `json:"wp_status_code"`
	DatabaseAccess        int     `json:"database_access"`
	PluginUpdateCount     int     `json:"plugin_update_count"`
	ThemeUpdateCount      int     `json:"theme_update_count"`
	CoreUpdateAvailable   bool    `json:"core_update_available"`
	UnapprovedComments    int     `json:"unapproved_comments"`
	ResponseTimeMs        float64 `json:"response_time_ms"`
	CurrentScriptMemoryMb float64 `json:"current_script_memory_mb"`
	PeakScriptMemoryMb    float64 `json:"peak_script_memory_mb"`
	WPVersion             string  `json:"wp_version"`
	PHPVersion            string  `json:"php_version"`
	DBQueryCount          int     `json:"db_query_count"`
}

func printUsage() {
	usageText := `
Usage: check_wp_status [OPTIONS]

This script monitors the status of a WordPress installation by querying an endpoint and checking various metrics.

OPTIONS:
  -H <url>          URL of the endpoint to check (required)
  -P <password>     Authorization password for header (optional)

Threshold parameters (optional):
  -Z <value>        Warn if core updates are above threshold (default: 0)
  -Y <value>        Warn if plugin updates are above threshold (default: 0)
  -X <value>        Warn if theme updates are above threshold (default: 0)
  -W <value>        Warn if unapproved comments are above threshold (default: 0)
  -V <value>        Warn if response time (ms) is above threshold (default: 0.0)
  -U <value>        Warn if peak memory usage (MB) is above threshold (default: 6.0)

General:
  -h, --help        Display this help message and exit
`
	fmt.Println(usageText)
}

func main() {
	// Parse flags for URL and thresholds
	flag.Usage = printUsage
	url := flag.String("H", "", "URL of the endpoint to check")
	password := flag.String("P", "", "Password for Authorization header (optional)")

	// Threshold flags
	coreUpdateThreshold := flag.Int("Z", 0, "Warn if core updates are above threshold")
	pluginUpdateThreshold := flag.Int("Y", 0, "Warn if plugin updates are above threshold")
	themeUpdateThreshold := flag.Int("X", 0, "Warn if theme updates are above threshold")
	unapprovedCommentsThreshold := flag.Int("W", 0, "Warn if unapproved comments are above threshold")
	responseTimeThreshold := flag.Float64("V", 1.0, "Warn if response time is above threshold")
	peakMemoryThreshold := flag.Float64("U", 12.0, "Warn if peak memory usage is above threshold")

	// Parse all flags
	flag.Parse()

	if *url == "" || *password == "" {
		fmt.Println("ERROR: The -H (URL) and -P (password) parameters are required")
		flag.Usage()
		os.Exit(3)
	}

	// Create HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", *url, nil)
	if err != nil {
		fmt.Printf("Error creating HTTP request: %s\n", err)
		os.Exit(2) // Critical
	}

	if *password != "" {
		req.Header.Add("Authorization", *password)
	}

	// Fetch JSON data
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching URL: %s\n", err)
		os.Exit(2) // Critical
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		os.Exit(2) // Critical
	}

	// Parse JSON data
	var wpStatus WPStatus
	err = json.Unmarshal(body, &wpStatus)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		os.Exit(2) // Critical
	}

	// Performance data string
	perfOutput := fmt.Sprintf("plugin_update_count=%d", wpStatus.PluginUpdateCount)
	perfOutput += fmt.Sprintf(" theme_update_count=%d", wpStatus.ThemeUpdateCount)
	perfOutput += fmt.Sprintf(" core_update_available=%t", wpStatus.CoreUpdateAvailable)
	perfOutput += fmt.Sprintf(" theme_update_count=%d", wpStatus.UnapprovedComments)
	perfOutput += fmt.Sprintf(" response_time_ms=%f", wpStatus.ResponseTimeMs)
	perfOutput += fmt.Sprintf(" peak_script_memory_mb=%f", wpStatus.PeakScriptMemoryMb)
	perfOutput += fmt.Sprintf(" wp_version=%s", wpStatus.WPVersion)
	perfOutput += fmt.Sprintf(" php_version=%s", wpStatus.PHPVersion)
	perfOutput += fmt.Sprintf(" db_query_count=%d", wpStatus.DBQueryCount)

	exitCode := 0
	outputMessage := ""
	if wpStatus.PluginUpdateCount > *pluginUpdateThreshold {
		outputMessage += fmt.Sprintf("Plugin updates exceed threshold (available: %d) ", wpStatus.PluginUpdateCount)
		exitCode = 1
	}
	if wpStatus.ThemeUpdateCount > *themeUpdateThreshold {
		outputMessage += fmt.Sprintf("Theme updates exceed threshold (available: %d) ", wpStatus.ThemeUpdateCount)
		exitCode = 1
	}
	if wpStatus.CoreUpdateAvailable && *coreUpdateThreshold == 0 {
		outputMessage += "Core update available "
		exitCode = 1
	}
	if wpStatus.UnapprovedComments > *unapprovedCommentsThreshold {
		outputMessage += fmt.Sprintf("%d Unapproved comments exceed %d threshold ", wpStatus.UnapprovedComments, *unapprovedCommentsThreshold)
		exitCode = 1
	}
	if wpStatus.ResponseTimeMs > *responseTimeThreshold {
		outputMessage += fmt.Sprintf("Response time (%.2f) exceeds threshold (%.2f) ", wpStatus.ResponseTimeMs, *responseTimeThreshold)
		exitCode = 1
	}
	if wpStatus.PeakScriptMemoryMb > *peakMemoryThreshold {
		outputMessage += fmt.Sprintf("Memory exceeds threshold (used: %.2f MB, threshold: %.2f MB) ", wpStatus.PeakScriptMemoryMb, *peakMemoryThreshold)
		exitCode = 1
	}

	// Final output
	if exitCode == 0 {
		fmt.Println("OK " + outputMessage + "| " + perfOutput)
		os.Exit(0) // OK in Nagios
	} else if exitCode == 1 {
		fmt.Println("WARNING " + outputMessage + "| " + perfOutput)
		os.Exit(1) // Warning in Nagios
	} else {
		fmt.Println("CRITICAL " + outputMessage + "| " + perfOutput)
		os.Exit(2) // Critical in Nagios
	}
}
