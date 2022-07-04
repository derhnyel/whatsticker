package push

import (
	"encoding/json"
	"log"

	rmq "github.com/adjust/rmq/v4"
	"github.com/derhnyel/whatsticker/metrics"
)

func (consumer *metrics.Register) Consume(delivery rmq.Delivery) {
	var stickerMetrics metrics.StickerizationMetric
	if err := json.Unmarshal([]byte(delivery.Payload()), &stickerMetrics); err != nil {
		// handle json error
		if err := delivery.Reject(); err != nil {
			// handle reject error
			log.Printf("Error delivering Reject %s", err)
		}
		return
	}

	metrics.CheckAndIncrementMetrics(stickerMetrics, &consumer.Gauges)
	metrics.PushToGateway(&consumer.Pusher)
	if err := delivery.Ack(); err != nil {
		// handle ack error
		return
	}

}
