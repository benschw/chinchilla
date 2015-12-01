package repeater

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/benschw/chinchilla/config"
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

	root := "/chinchilla/repeater/connections/"
	kv := client.KV()

	results, _, err := kv.List(root, nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	arr := make([]config.RabbitAddress, 0)
	for _, p := range results {
		if p.Key == root {
			continue
		}
		add := &config.RabbitAddress{}

		if err = yaml.Unmarshal(p.Value, add); err != nil {
			log.Printf("Error Unmarshaling EP Config: %s", err)
			continue
		}
		arr = append(arr, *add)
	}

	repLib := NewRepeaterLib(arr)

	// Register Consumer and Delivery strategies
	ep.RegisterDeliveryStrategy(TopicRepeaterStrategy, &Repeater{Lib: repLib})
}
