package beater

import (
	//"bufio"
	"bytes"
	//"fmt"
	"io/ioutil"
	"log" // how do I do logging from beats? May be their logp?
	//"path/filepath"
	"strconv"
	//"strings"
)

// ZFS will use maps for stats because I do not want to be specific about fields

func GetZfsPoolIofile(s string) (result map[string]uint64, err error) {
	buf, err := ioutil.ReadFile(s)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(buf))
	// first line is something ZFS specific with a timestamp, second is names, third is valus and fourth empty
	// using splits for just two lines?
	bb := bytes.Split(buf, []byte("\n"))
	if len(bb) > 2 {
		keys := bytes.Fields(bb[1])
		svals := bytes.Fields(bb[2])
		l := len(keys)
		if len(svals) != l {
			log.Fatal("Number of fields mismatch", keys, svals)
		}
		result = make(map[string]uint64)
		for i, k := range keys {
			v, err := strconv.ParseUint(string(svals[i]), 10, 64)
			if err != nil {
				log.Fatal("Parse failure in zfs io", err)
			}
			result[string(k)] = v
		}
		//fmt.Println(result)
	} //else {
	//	fmt.Println("strange io file, len", len(bb))
	//}
	return result, err
}

func GetZfsStatfile(s string) (result map[string]uint64, err error) {
	// reads one of ZFS stats files from kstats and returns everything as a map
	err = nil
	buf, err := ioutil.ReadFile(s)
	if err != nil {
		log.Fatal(err)
	}
//	fmt.Println(string(buf))

	bb := bytes.Split(buf, []byte("\n"))
	if len(bb) < 2 {
		log.Fatal("invalid len is:: ", len(bb))
	}
	result = make(map[string]uint64) // or is it 32? what is the "type"?

	for _, bbb := range bb[2:] {
		//fmt.Println( len(bbb), string(bbb))
		ll := bytes.Fields(bbb)
		if len(ll) != 3 {
			//only expecting name 4  value there, discarding the 4
			break
		}
		v, _ := strconv.ParseUint(string(ll[2]), 10, 64) // are they always positive? I hope so that uint works
		result[string(ll[0])] = v
	}
	return result, err
}

