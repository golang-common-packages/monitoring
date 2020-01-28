package monitoring

import (
	"github.com/labstack/echo/v4"
)

// IMonitoring interface for monitoring package
type IMonitoring interface {
	Middleware() echo.MiddlewareFunc
}

const (
	NEWRELIC = iota
	PGO
)

// New function for Factory Pattern
func New(monitoringType int, serviceName, licenseKey string) IMonitoring {
	switch monitoringType {
	case NEWRELIC:
		return NewRelic(serviceName, licenseKey)
	case PGO:
		return NewPGO(serviceName)
	}

	return nil
}
