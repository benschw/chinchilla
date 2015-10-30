package queue

import (
	"time"

	"github.com/streadway/amqp"
)

var conn *amqp.Connection

func init() {
	c, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	conn = c
}
func countMessages(msgs <-chan amqp.Delivery) int {

	var cnt = 0
	for {
		select {
		case d, _ := <-msgs:
			d.Ack(false)
			if d.Body == nil {
				return cnt
			}
			cnt++
		case <-time.After(5 * time.Millisecond):
			return cnt
		}
	}
	return 0
}
