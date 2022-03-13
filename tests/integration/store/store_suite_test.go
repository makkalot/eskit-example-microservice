package store

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/makkalot/eskit-example-microservice/services/clients"
	"golang.org/x/net/context"
	"os"
	"testing"
)

func TestUsers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Store Suite")
}

var (
	storeEndpoint     string
	crudStoreEndpoint string
)

var _ = BeforeSuite(func() {

	storeEndpoint = os.Getenv("EVENTSTORE_ENDPOINT")
	if storeEndpoint == "" {
		Fail("EVENTSTORE_ENDPOINT is required")
	}

	crudStoreEndpoint = os.Getenv("CRUDSTORE_ENDPOINT")
	if crudStoreEndpoint == "" {
		Fail("CRUDSTORE_ENDPOINT is required")
	}

	waitForPreReqServices()
})

func waitForPreReqServices() {
	_, err := clients.NewStoreClientWithWait(context.Background(), storeEndpoint)
	if err != nil {
		Fail(fmt.Sprintf("couldn't connect to %s because of %v", storeEndpoint, err))
	}
}
