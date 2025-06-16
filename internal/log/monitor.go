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

// MonitorLogs retrieves and streams logs from containers in the specified Docker network.
func MonitorLogs(ctx context.Context, cli *client.Client, networkName string) error {
	network, err := cli.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to find network '%s': %v", networkName, err)
	}

	containers := network.Containers
	if len(containers) == 0 {
		fmt.Printf("No containers found in network '%s'.\n", networkName)
		return nil
	}

	fmt.Printf("Monitoring logs for network '%s'...\n", networkName)

	for containerID, containerInfo := range containers {
		go streamContainerLogs(ctx, cli, containerID, containerInfo.Name)
	}

	return nil
}

// streamContainerLogs streams logs from a specific container and formats them with timestamps and colors.
func streamContainerLogs(ctx context.Context, cli *client.Client, containerID, containerName string) {
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