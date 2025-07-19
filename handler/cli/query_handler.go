package cli

import (
	"fmt"
	"strings"
)

type QueryHandler struct {
	renderer *FormatRenderer
}

func NewQueryHandler(renderer *FormatRenderer) *QueryHandler {
	return &QueryHandler{
		renderer: renderer,
	}
}

func (h *QueryHandler) HandleFormatQuery(formatString string) error {
	result, err := h.processFormat(formatString)
	h.outputResult(result, err)
	return err
}

func (h *QueryHandler) processFormat(formatString string) (string, error) {
	// Use FormatRenderer to handle variable substitution
	return h.renderer.Render(formatString)
}

func (h *QueryHandler) outputResult(result string, err error) {
	if err != nil {
		// Output consistent error message for all failure scenarios
		// This provides graceful degradation as specified in requirements
		fmt.Print("❌ ERROR")
	} else {
		fmt.Print(result)
	}
}

// isConnectionError checks if the error is related to connection issues
func (h *QueryHandler) isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := strings.ToLower(err.Error())

	// Check for common connection error patterns
	connectionPatterns := []string{
		"connection refused",
		"connection timeout",
		"context deadline exceeded",
		"network is unreachable",
		"no such host",
		"failed to connect",
		"server is unavailable",
		"dial",
		"grpc",
	}

	for _, pattern := range connectionPatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

// isTimeoutError checks if the error is specifically a timeout
func (h *QueryHandler) isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "context deadline exceeded") ||
		strings.Contains(errorMsg, "timeout")
}

// isConfigurationError checks if the error is related to configuration issues
func (h *QueryHandler) isConfigurationError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := strings.ToLower(err.Error())

	// Check for configuration-related error patterns
	configPatterns := []string{
		"failed to read config",
		"invalid configuration",
		"missing configuration",
		"config not found",
	}

	for _, pattern := range configPatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}
