package log

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
)

// LogConfig holds the configuration for log monitoring.
type LogConfig struct {
	serviceNames []string
}

// streamContainerLogs streams logs from a specific container and formats them with timestamps and colors.
func streamContainerLogs(ctx context.Context, cli *client.Client, containerID, containerName string, config *LogConfig) {
	logs, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		fmt.Printf("Error streaming logs for %s: %v\n", containerName, err)
		return
	}
	defer logs.Close()

	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		logLine := scanner.Text()
		if logLine == "" {
			continue
		}

		timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
		logColor := getLogColor(logLine)

		logColor.Printf("[%s] [%s] %s\n", timestamp, containerName, logLine)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading logs for %s: %v\n", containerName, err)
	}
}

// getLogColor determines the appropriate color for the log line based on its content.
func getLogColor(logLine string) *color.Color {
	logLineLower := strings.ToLower(logLine)
	if strings.Contains(logLineLower, "error") {
		return color.New(color.FgRed, color.Bold)
	} else if strings.Contains(logLineLower, "warn") {
		return color.New(color.FgYellow, color.Bold)
	}
	return color.New(color.FgGreen)
}
