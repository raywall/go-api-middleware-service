// Package main contains the metrics collection logic for the record query microservice.
package main

import (
	"fmt"

	"github.com/DataDog/datadog-go/statsd"

	_ "github.com/go-sql-driver/mysql"
)

// metric_name defines the name of the metric for request counting.
const metric_name string = "registros.requests_total"

// statsdClient holds the global Datadog statsd client for metrics.
var (
	statsdClient *statsd.Client
)

// registrarMetrica increments a Datadog metric with the provided tags.
func registrarMetrica(tags []string) error {
	// Increment the metric with the specified tags and a value of 1.
	if err := statsdClient.Incr(metric_name, tags, 1); err != nil {
		// Return a wrapped error if the metric increment fails.
		return fmt.Errorf("erro ao registrar métrica registros.requests_total: %w", err)
	}
	// Return no error on success.
	return nil
}
