package monitoring

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	nr "github.com/newrelic/go-agent"
)

// NewRelicClient manage all Relic action
type NewRelicClient struct {
	Session    nr.Application
	LicenseKey string
}

const (
	NEWRELIC_TXN = "newrelic-txn"
)

// NewRelic function return a new relic client based on singleton pattern
func NewRelic(serviceName, licenseKey string) IMonitoring {
	configRelic := nr.NewConfig(serviceName, licenseKey)
	app, err := nr.NewApplication(configRelic)
	if err != nil {
		panic(fmt.Errorf("New relic: %s", err))
	}

	currentSession := &NewRelicClient{app, licenseKey}
	log.Println("Connected to Relic Server")

	return currentSession
}

// Middleware returns a middleware that collect request data for NewRelic
func (n *NewRelicClient) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if n.Session == nil {
				next(c)
				return nil
			}

			transactionName := fmt.Sprintf("%s [%s]", c.Path(), c.Request().Method)
			txn := n.Session.StartTransaction(transactionName, c.Response().Writer, c.Request())
			defer txn.End()

			c.Set(NEWRELIC_TXN, txn)
			err := next(c)
			if err != nil {
				txn.NoticeError(err)
			}

			return err
		}
	}
}
