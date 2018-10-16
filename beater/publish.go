package beater

import (
	"time"
	"strings"
	"path/filepath"
	"log"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	//"github.com/gshamov/lustrebeat/config"
)


func PublishZpool(bt *Lustrebeat, v string) (err error) {
	// publishes an event for ZFS pool/io

	// make the tag
	pool := strings.Split(v, "/")[5]
	ios, err := GetZfsPoolIofile(v)
	if err != nil {
		//log.Fatal(err) // is in fatal here?
		return err
	}
	event := beat.Event{
		Timestamp: time.Now(),
		Fields: common.MapStr{
			"type": "zfsio",
			"ios":  ios,
			"pool": pool,
		},
	}
	bt.client.Publish(event)
	logp.Info("Event sent")
	return err
}

func PublishZstats(bt *Lustrebeat, s string, listzstats map[string]struct{}) (err error) {
	// publishes a ZFS stat file, based on the mask?

	cs := filepath.Base(s)
	var ok bool = false
	// no provision for empty lists! should be explicit stats always
	// or nothing will be returned

	if _, ok = listzstats[cs]; ok {
		result, err := GetZfsStatfile(s)
		//fmt.Println(result, err, s)
		if err != nil {
			log.Fatal(err)
			return err
		}
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    "zfsstat",
				"zfsstat": result,
				"name":    cs,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
	}
	return err
}
