package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
)

// Colors for log output
var (
	normalColor  = color.New(color.FgGreen)
	warningColor = color.New(color.FgYellow, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
)

// LogConfig holds the logging configuration
type LogConfig struct {
	showAll      bool
	showErrors   bool
	showWarns    bool
	showInfo     bool
	showDebug    bool
	customWords  string
	logLevels    string
	serviceNames []string // Change from string to []string
}

func main() {
	// Parse CLI arguments
	networkName := flag.String("network", "", "Docker network name to monitor")
	config := LogConfig{}
	flag.BoolVar(&config.showAll, "all", false, "Show all logs (default: false)")
	flag.BoolVar(&config.showErrors, "errors", false, "Show error logs (default: false)")
	flag.BoolVar(&config.showWarns, "warnings", false, "Show warning logs (default: false)")
	flag.BoolVar(&config.showInfo, "info", false, "Show info logs (default: false)")
	flag.BoolVar(&config.showDebug, "debug", false, "Show debug logs (default: false)")
	flag.StringVar(&config.customWords, "filter", "", "Additional keywords to filter, comma-separated")
	flag.StringVar(&config.logLevels, "levels", "", "Comma-separated log levels to show (error,warn,info,debug)")
	var serviceList string
	flag.StringVar(&serviceList, "service", "", "Filter logs by service names (comma-separated, partial match)")

	flag.Parse()

	// If no specific level is selected, show all
	if !config.showErrors && !config.showWarns && !config.showInfo && !config.showDebug && config.logLevels == "" {
		config.showAll = true
	}

	// Parse log levels if specified
	if config.logLevels != "" {
		levels := strings.Split(strings.ToLower(config.logLevels), ",")
		for _, level := range levels {
			switch strings.TrimSpace(level) {
			case "error":
				config.showErrors = true
			case "warn", "warning":
				config.showWarns = true
			case "info":
				config.showInfo = true
			case "debug":
				config.showDebug = true
			}
		}
	}

	// Parse service names if specified
	if serviceList != "" {
		config.serviceNames = parseServiceNames(serviceList)
	}

	if *networkName == "" {
		fmt.Println("Error: --network is required")
		flag.Usage()
		os.Exit(1)
	}

	// Set up Docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Error initializing Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("\nStopping log monitor...")
		os.Exit(0)
	}()

	// Monitor logs
	if err := monitorLogs(ctx, cli, *networkName, &config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func monitorLogs(ctx context.Context, cli *client.Client, networkName string, config *LogConfig) error {
	// Get network details
	network, err := cli.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to find network '%s': %v", networkName, err)
	}

	// Get containers in the network
	containers := network.Containers
	if len(containers) == 0 {
		fmt.Printf("No containers found in network '%s'.\n", networkName)
		return nil
	}

	fmt.Printf("Monitoring logs for network '%s'...\n", networkName)

	// Use WaitGroup to keep main function alive until all logs are done
	var wg sync.WaitGroup
	for containerID, containerInfo := range containers {
		wg.Add(1)
		go func(id, name string) {
			defer wg.Done()
			streamContainerLogs(ctx, cli, id, name, config)
		}(containerID, containerInfo.Name)
	}

	// Wait for all goroutines to finish (they won't unless interrupted)
	wg.Wait()
	return nil
}

func streamContainerLogs(ctx context.Context, cli *client.Client, containerID, containerName string, config *LogConfig) {
	// Get container details to fetch service name (if using Docker Compose)
	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		fmt.Printf("Error inspecting container %s: %v\n", containerName, err)
		return
	}

	// Use service name from Docker Compose labels, fallback to container name
	serviceName := containerName
	if labels, exists := containerInfo.Config.Labels["com.docker.compose.service"]; exists {
		serviceName = labels
	}

	// Check if container name matches the service filter
	if !matchesServiceName(containerName, config.serviceNames) {
		return
	}

	// Stream logs
	logs, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		fmt.Printf("Error streaming logs for %s: %v\n", serviceName, err)
		return
	}
	defer logs.Close()

	// Read logs line by line
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		logLine := scanner.Text()
		if logLine == "" {
			continue
		}

		// Sanitize the log line
		logLine = sanitizeLogLine(logLine)
		if logLine == "" {
			continue
		}

		logLineLower := strings.ToLower(logLine)
		if !shouldLogMessage(logLineLower, config) {
			continue
		}

		// Format timestamp
		timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")

		// Determine log style based on content
		var logColor *color.Color
		if isErrorMessage(logLineLower) {
			logColor = errorColor
		} else if isWarningMessage(logLineLower) {
			logColor = warningColor
		} else {
			logColor = normalColor
		}

		// Print log with timestamp and service name
		logColor.Printf("[%s] [%s] %s\n", timestamp, serviceName, logLine)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading logs for %s: %v\n", serviceName, err)
	}
}

func shouldLogMessage(logLine string, config *LogConfig) bool {
	if config.showAll {
		return true
	}

	// Check custom keywords first
	if config.customWords != "" {
		customKeywords := strings.Split(config.customWords, ",")
		for _, keyword := range customKeywords {
			if strings.Contains(logLine, strings.TrimSpace(strings.ToLower(keyword))) {
				return true
			}
		}
	}

	// Check log levels
	if config.showErrors && isErrorMessage(logLine) {
		return true
	}
	if config.showWarns && isWarningMessage(logLine) {
		return true
	}
	if config.showInfo && isInfoMessage(logLine) {
		return true
	}
	if config.showDebug && isDebugMessage(logLine) {
		return true
	}

	return false
}

// Add these helper functions to replace the existing ones
func isErrorMessage(logLine string) bool {
	// Check if "Errors: []" or "error: null" appears, indicating no actual errors
	if strings.Contains(logLine, "errors: []") ||
		strings.Contains(logLine, "errors:[]") ||
		strings.Contains(logLine, "error: null") {
		return false
	}

	errorPatterns := []struct {
		keyword string
		context string
	}{
		{"error", ""},
		{"exception", ""},
		{"failed", "failure"},
		{"panic", ""},
		{"fatal", ""},
		{"critical", ""},
	}

	for _, pattern := range errorPatterns {
		if pattern.context == "" {
			// Simple keyword match
			if strings.Contains(logLine, pattern.keyword) {
				// Make sure it's not part of a "no error" or "error: null" message
				if !strings.Contains(logLine, "no "+pattern.keyword) &&
					!strings.Contains(logLine, pattern.keyword+": null") {
					return true
				}
			}
		} else {
			// Contextual match
			if strings.Contains(logLine, pattern.keyword) &&
				strings.Contains(logLine, pattern.context) {
				return true
			}
		}
	}
	return false
}

func isWarningMessage(logLine string) bool {
	// Skip if it's a status change or success message
	if strings.Contains(logLine, "status from") ||
		strings.Contains(logLine, "changed status") ||
		strings.Contains(logLine, "success") {
		return false
	}

	warningKeywords := []string{
		"warn",
		"warning",
		"deprecated",
		"timeout",
		"unavailable",
	}

	// Only check "retry" if it's accompanied by an error context
	if strings.Contains(logLine, "retry") &&
		(strings.Contains(logLine, "failed") ||
			strings.Contains(logLine, "error") ||
			strings.Contains(logLine, "timeout")) {
		return true
	}

	for _, keyword := range warningKeywords {
		if strings.Contains(logLine, keyword) {
			return true
		}
	}
	return false
}

func isInfoMessage(logLine string) bool {
	infoKeywords := []string{
		"info",
		"information",
		"notice",
		"success",
	}
	for _, keyword := range infoKeywords {
		if strings.Contains(logLine, keyword) {
			return true
		}
	}
	return false
}

func isDebugMessage(logLine string) bool {
	debugKeywords := []string{
		"debug",
		"trace",
		"verbose",
	}
	for _, keyword := range debugKeywords {
		if strings.Contains(logLine, keyword) {
			return true
		}
	}
	return false
}

// Add a new helper function to check service name match
func matchesServiceName(containerName string, serviceFilters []string) bool {
	if len(serviceFilters) == 0 {
		return true
	}

	containerNameLower := strings.ToLower(containerName)
	for _, filter := range serviceFilters {
		if strings.Contains(containerNameLower, strings.ToLower(filter)) {
			return true
		}
	}
	return false
}

// Add a new helper function to parse service names
func parseServiceNames(serviceList string) []string {
	names := strings.Split(serviceList, ",")
	var filtered []string
	for _, name := range names {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return filtered
}

// Add this function after the existing imports
func sanitizeLogLine(logLine string) string {
	// Remove common control characters and invalid UTF-8 sequences
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, logLine)

	// Handle JSON-like structures
	sanitized = strings.ReplaceAll(sanitized, "\u0000", "")

	// Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	sanitized = ansiRegex.ReplaceAllString(sanitized, "")

	return strings.TrimSpace(sanitized)
}
