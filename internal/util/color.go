package util

import "github.com/fatih/color"

// Colors for log output
var (
	NormalColor  = color.New(color.FgGreen)
	WarningColor = color.New(color.FgYellow, color.Bold)
	ErrorColor   = color.New(color.FgRed, color.Bold)
)

// ColorizeLog formats the log message with the appropriate color based on its level.
func ColorizeLog(level string, message string) string {
	switch level {
	case "error":
		return ErrorColor.Sprint(message)
	case "warn":
		return WarningColor.Sprint(message)
	default:
		return NormalColor.Sprint(message)
	}
}