package provider

import (
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/eventstore"
	common2 "github.com/makkalot/eskit/lib/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
)

type eventAction string

const (
	created eventAction = "created"
	deleted eventAction = "deleted"
	updated eventAction = "updated"
)

var (
	eventStoreEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eskit_events_store_total",
			Help: "Total events in the eventstore",
		}, []string{
			"entity_type", "event_type",
		})

	crudStoreEvents = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eskit_events_crud_total",
			Help: "Total events for the crudstore",
		}, []string{
			"entity_type", "event_type",
		})
)

type PrometheusMetricsConsumer struct {
	ctx context.Context
}

func NewPrometheusMetricsConsumer(ctx context.Context) *PrometheusMetricsConsumer {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":8888", nil)
	}()

	return &PrometheusMetricsConsumer{
		ctx: ctx,
	}
}

func (consumer *PrometheusMetricsConsumer) ConsumerCB(entry *eventstore.AppLogEntry) error {
	//log.Println("Consuming Metrics event : ", entry.Event.EventType)
	entityType := common2.ExtractEntityType(entry.Event)
	eventType := common2.ExtractEventType(entry.Event)

	eventsCounter := eventStoreEvents.With(prometheus.Labels{"entity_type": entityType, "event_type": eventType})
	eventsCounter.Inc()

	if consumer.isCrudEvent(entry.Event) {
		crudCounter := crudStoreEvents.With(prometheus.Labels{"entity_type": entityType, "event_type": eventType})
		crudCounter.Inc()
	}

	return nil
}

func (consumer *PrometheusMetricsConsumer) isCrudEvent(event *eventstore.Event) bool {
	eventType := common2.ExtractEventType(event)
	eventType = strings.ToLower(eventType)

	switch eventAction(eventType) {
	case created, updated, deleted:
	default:
		return false
	}

	return true
}
