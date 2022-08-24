package util

import (
	"fmt"
	"strings"
)

func GetEnvLogsTableName(envID string) string {
	return fmt.Sprintf("logs_%s", strings.ReplaceAll(envID, "-", "_"))
}

func GetEnvEventsTableName(envID string) string {
	return fmt.Sprintf("events_%s", strings.ReplaceAll(envID, "-", "_"))
}
