package beater

import (
	"time"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	//"github.com/gshamov/lustrebeat/config"
)


func Counter(bt *Lustrebeat, t string, counter uint64 ) {
	// test fuction yo 
    		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    t,
				"counter2": counter,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent from Counter")
}