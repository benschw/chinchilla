package repeater

import (
	"log"
	"os"

	"github.com/benschw/chinchilla/ep"
	"github.com/hashicorp/consul/api"
)

const TopicRepeaterStrategy = "topic-repeater"

func init() {

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	conProvider := NewConnectionProvider(&ConsulConAddressProvider{
		Root:   "/chinchilla",
		Client: client,
	})

	repLib := NewRepeaterLib(conProvider)

	// Register Consumer and Delivery strategies
	ep.RegisterDeliveryStrategy(TopicRepeaterStrategy, &Repeater{Lib: repLib})
}
