package main

import (
	"context"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/makkalot/eskit-example-microservice/services/clients"
	metrics "github.com/makkalot/eskit-example-microservice/services/consumers/metrics/provider"
	"github.com/makkalot/eskit/lib/consumer"
	"github.com/makkalot/eskit/lib/consumerstore"
	"github.com/spf13/viper"
	"log"
)

type ConsumerConfig struct {
	DbUri              string `json:"dbUri" mapstructure:"dbUri"`
	ConsumerName       string `json:"consumerName" mapstructure:"consumerName"`
	CrudStoreEndpoint  string `json:"crudStoreEndpoint" mapstructure:"crudStoreEndpoint"`
	EventStoreEndpoint string `json:"eventStoreEndpoint" mapstructure:"eventStoreEndpoint"`
}

func (c ConsumerConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ConsumerName, validation.Required),
		validation.Field(&c.DbUri, validation.Required),
		validation.Field(&c.EventStoreEndpoint, validation.Required),
		validation.Field(&c.CrudStoreEndpoint, validation.Required),
	)
}

func main() {
	viper.BindEnv("consumerName", "CONSUMER_NAME")
	viper.BindEnv("eventStoreEndpoint", "EVENT_STORE_ENDPOINT")
	viper.BindEnv("crudStoreEndpoint", "CRUDSTORE_ENDPOINT")
	viper.BindEnv("dbURI", "DB_URI")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/metrics")
	viper.AddConfigPath(".")

	var config ConsumerConfig

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("config validation : %v", err)
	}

	ctx := context.Background()

	eventStoreClient, err := clients.NewStoreClient(ctx, config.EventStoreEndpoint)
	if err != nil {
		log.Fatalf("event store client initialization failed : %v", err)
	}

	consumeStore, err := consumerstore.NewSQLConsumerApiProvider(config.DbUri)
	if err != nil {
		log.Fatalf("consumer store initialization failed : %v", err)
	}

	metricsConsumer := metrics.NewPrometheusMetricsConsumer(ctx)

	appLogConsumer, err := consumer.NewAppLogConsumer(eventStoreClient, consumeStore, config.ConsumerName, consumer.FromSaved, "*")
	if err != nil {
		log.Fatalf("applog consumer initialization failed : %v", err)
	}

	if err := appLogConsumer.Consume(ctx, metricsConsumer.ConsumerCB); err != nil {
		log.Fatalf("consuming failed : %v", err)
	}
}
