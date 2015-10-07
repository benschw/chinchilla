package queue

import "github.com/benschw/chinchilla/ep"

func init() {
	ep.RegisterQueueType(
		ep.DefaultQueueType,
		&Queue{C: &DefaultWorker{}, D: &DefaultDeliverer{}},
	)
}
