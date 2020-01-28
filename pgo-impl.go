package monitoring

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"github.com/labstack/echo/v4"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// PGOClient manage all PGO action
type PGOClient struct {
	Handler *prometheus.Exporter
}

var (
	Latency              = stats.Float64("repl/latency", "Latency", "ms")
	totalRequestAccepted = stats.Int64("repl/totalRequestAccepted", "Total request", "ms")
	errorCount           = stats.Int64("repl/errorCount", "Error count", "ms")

	keyMethod, _ = tag.NewKey("method")
	keyStatus, _ = tag.NewKey("status")
	keyError, _  = tag.NewKey("error")
	keyPath, _   = tag.NewKey("path")

	LatencyView = &view.View{
		Name:        "api/latency",
		Measure:     Latency,
		Description: "Latency",
		Aggregation: view.Distribution(),
		TagKeys:     []tag.Key{keyMethod, keyPath, keyStatus, keyError}}

	totalRequest = &view.View{
		Name:        "api/totalRequestAccepted",
		Measure:     totalRequestAccepted,
		Description: "Total request",
		Aggregation: view.Count(),
	}
)

// NewPGO function return a exporter client based on singleton pattern
func NewPGO(serviceName string) IMonitoring {
	// if err := view.Register(LatencyView, BytesInView, BytesOutView); err != nil {
	if err := view.Register(LatencyView, totalRequest); err != nil {

		panic(err)
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: serviceName,
	})
	if err != nil {
		panic(err)
	}

	view.RegisterExporter(pe)

	currentSession := &PGOClient{pe}
	log.Println("Connected to PGO Server")

	return currentSession
}

// Middleware returns a middleware that collect request data for PGO
func (p *PGOClient) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if p == nil {
				next(c)
				return nil
			}
			start := time.Now()
			err := next(c)
			stop := time.Now()
			l := float64(stop.Sub(start))
			if err != nil {
				ctx, err := tag.New(context.Background(), tag.Insert(keyMethod, c.Request().Method), tag.Insert(keyStatus, err.Error()), tag.Insert(keyPath, c.Path()))
				if err != nil {
					return err
				}

				stats.Record(ctx, Latency.M(l))
			} else {
				ctx, err := tag.New(context.Background(), tag.Insert(keyMethod, c.Request().Method), tag.Insert(keyStatus, fmt.Sprintf("%d", c.Response().Status)), tag.Insert(keyPath, c.Path()))
				if err != nil {
					return err
				}
				bytesIn, _ := strconv.ParseInt(c.Request().Header.Get(echo.HeaderContentLength), 10, 64)
				stats.Record(ctx, Latency.M(l), totalRequestAccepted.M(bytesIn), errorCount.M(c.Response().Size))
			}

			return err
		}
	}
}
