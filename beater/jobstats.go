package beater

import (
	//"bytes"
	"fmt"
	"io/ioutil"
	//"log" // how do I do logging from beats? May be their logp?
	"path/filepath"

	//"strings"
	"time"
	// for yaml-bases parsing of jobstats
	"gopkg.in/yaml.v2"
	"regexp"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	//"github.com/gshamov/lustrebeat/config"
)

// recursive function to deal with Yaml data
func Convert(i interface{}) interface{} {
	// lifted from the StackOverflow sample. It works.
	// https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = Convert(v)
		}
		return m2
		/*	case map[string]interface{}:
			m2 := map[string]interface{}{}
			for k, v := range x {
				m2[k.(string)] = Convert(v)
			}
			return m2 */
	case []interface{}:
		for i, v := range x {
			x[i] = Convert(v)
		}
	}
	return i
}

func JobStat(bt *Lustrebeat, s string) (err error) {
	// pars a single jobstat file s using GoYaml magic, then publishes into BT

	re := regexp.MustCompile(`([a-z]+\:)`) // regexp for jobstats max:12345 out of the loop

	buf, err := ioutil.ReadFile(s)
	if err != nil {
		//log.Fatal(err)
		return err
	}

	fs, obj, _, _ := MakeLustreTagsE(filepath.Dir(s)) // for tagging, common to all records below

	buf1 := re.ReplaceAll(buf, []byte("$1 ")) // adding spaces for jammed strings like max:123456
	buf = nil
	//fmt.Println(string(buf1))

	// need manual recursion because map[]interface{} is useless
	// code from https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang

	var data interface{}

	err = yaml.Unmarshal(buf1, &data)
	if err != nil {
		//log.Fatal(err)
		return err
	}

	data = Convert(data)

	// unholy magic of Golang dynamic maps from https://blog.golang.org/json-and-go

	m := data.(map[string]interface{})
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			fmt.Println(k, "is string", vv)
		case float64:
			fmt.Println(k, "is float64", vv)
		case []interface{}:
			fmt.Println(k, "is an array:")
			for _, u := range vv {
				w := u.(map[string]interface{})

				event_jobstat := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":    "jobstat",
						"jobstat": w,
						"fs":      fs,
						"object":  obj,
					},
				}
				bt.client.Publish(event_jobstat)
				logp.Info("A jobstat record sent")

			}
		default:
			fmt.Println("unknown type")
		}
	}
	
	return nil
}
