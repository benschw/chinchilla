package queue

import "github.com/benschw/chinchilla/ep"

const FanOutStrategy = "fanout"

func init() {
	// Register Consumer and Delivery strategies
	ep.RegisterConsumerStrategy(ep.DefaultConsumerStrategy, &DefaultWorker{})
	ep.RegisterDeliveryStrategy(ep.DefaultDeliveryStrategy, &DefaultDeliverer{})
	ep.RegisterConsumerStrategy(FanOutStrategy, &FanOut{})
}
