package ep

import "fmt"

// Registry to hold all available Consume and Deliver implementations
type StrategyRegistry struct {
	DefaultConsumerStrategy string
	DefaultDeliveryStrategy string
	consumerReg             map[string]MsgConsumer
	deliveryReg             map[string]MsgDeliverer
}

// Set up in init
var strategyReg *StrategyRegistry

// Register a new Consumer implementation with the StrategyRegistry
func RegisterConsumerStrategy(key string, s MsgConsumer) {
	strategyReg.consumerReg[key] = s
}

// Register a new Deliverer implementation with the StrategyRegistry
func RegisterDeliveryStrategy(key string, s MsgDeliverer) {
	strategyReg.deliveryReg[key] = s
}

// Build a Strategy composition based on configuration
func GetStrategy(cKey string, dKey string) (*Strategy, error) {
	if cKey == "" {
		cKey = strategyReg.DefaultConsumerStrategy
	}
	if dKey == "" {
		dKey = strategyReg.DefaultDeliveryStrategy
	}
	consumer, ok := strategyReg.consumerReg[cKey]
	if !ok {
		return nil, fmt.Errorf("Consumer strategy labeled '%s' doesn't exist", cKey)
	}
	deliverer, ok := strategyReg.deliveryReg[dKey]
	if !ok {
		return nil, fmt.Errorf("Delivery strategy labeled '%s' doesn't exist", dKey)
	}
	return &Strategy{C: consumer, D: deliverer}, nil
}

func init() {
	// Initialize the Strategy Registry so other packages can start registering strategies
	strategyReg = &StrategyRegistry{
		DefaultConsumerStrategy: DefaultConsumerStrategy,
		DefaultDeliveryStrategy: DefaultDeliveryStrategy,
		consumerReg:             make(map[string]MsgConsumer),
		deliveryReg:             make(map[string]MsgDeliverer),
	}
}
