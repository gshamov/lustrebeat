package beater

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	//"log" // how do I do logging from beats? May be their logp?
	"path/filepath"
	"strconv"
	"strings"
)

//  I have decided to have two structs to supress lots of zeroes for single-val metrics

// For now, lets drop also SumSquared for it is very large and perhaps needs specal handling
// Also, Units are commented out for saving the space! We can get this from the name? Or document around it?

type LustreStatsRecordLong struct { // structs or maps?
	Metric  string
	Samples uint64
	//Unit    string
	Min     uint64
	Max     uint64
	Sum     uint64
	//	SumSquared float64
}

type LustreStatsRecordShort struct { // structs or maps?
	Metric  string
	Samples uint64
	//Unit    string
}

func MakeLustreTagsE(s string) (fs string, object string, client string, ok bool) {
	// tries to get Lustre FS and {MDT,OST, mounthash} from the path s
	// also will try to get the client name if that is to be expected
	//  all the pathes should be under /proc/fs/lustre ?

	if ok, _ := filepath.Match("/proc/fs/lustre/*/*", s); ok {
		// a lustre stats record, should have enough path by match
		zz := strings.Split(s, "/")
		kk := strings.Split(zz[5], "-")
		if len(kk) >= 2 {
			return kk[0], kk[1], "", true
		}
	}
	if ok, _ := filepath.Match("/proc/fs/lustre/*/*/exports/*o2ib", s); ok {
		// an exports stats record
		zz := strings.Split(s, "/")
		kk := strings.Split(zz[5], "-")
		if (len(kk) >= 2) && (len(zz) >= 8) {
			return kk[0], kk[1], zz[7], true
		}
	}
	// wtf happened?
	return "", "", "", false
}

func GetStatsFile1(s string) (result map[string]interface{}, err error) {
	// this is the new variant that uses new parsing routine below

	//GAS debug
	//fmt.Println("from GetStatsFile1", s)
	//GAS debug

	// here, or later, do some smarted tagging and save to the struct.

	buf, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	
	//fmt.Println(string(buf))
	
	result = make(map[string]interface{})
	
	// this is good for debug but makes a lot of space. Also forces interface{}, along with the units
	//result["procname"] = s

	// lets loop through the lines and populate the map
	b := bufio.NewScanner(bytes.NewBuffer(buf)) // am I being an idiot here? Probably; this reader works though
	for b.Scan() {
		bs := b.Text()
		ok := proc2slice(bs, result) // result is a map and thus goes by reference, will be changed
		if !ok {
			fmt.Println("error parsing line:", bs)
		}
	}

	return result, nil
}

func proc2slice(bs string, m map[string]interface{}) (ok bool) {
	// takes a string from stats file, parses it and stores in the map m

	as := bytes.Fields([]byte(bs)) // strings.Fields do not work , unicode?

	var lsl LustreStatsRecordLong
	var lss LustreStatsRecordShort
	var err error
	ok = false

	l := len(as)
	if l == 3 {
		// seconds to be skipped
		//lsr.Name = string(as[0])
		//lsr.SnapshotSec , _ = strconv.ParseFloat(string(as[1]),64)
		//fmt.Println(l, as, 3)
		return true
	} else {
		if l >= 4 {
			lss.Metric = string(as[0])
			lss.Samples, err = strconv.ParseUint(string(as[1]), 10, 64)
			if err != nil {
				fmt.Println(err)
				return false
			}
			//lss.Unit = string(as[3])
			m[lss.Metric] = lss
			//fmt.Println("lss", lss)
		}
		if l >= 7 {
			// copy over instead of reparsing
			lsl.Samples = lss.Samples
			//lsl.Unit = lss.Unit
			lsl.Metric = lss.Metric
			lsl.Min, err = strconv.ParseUint(string(as[4]), 10, 64)
			lsl.Max, err = strconv.ParseUint(string(as[5]), 10, 64)
			lsl.Sum, err = strconv.ParseUint(string(as[6]), 10, 64)
			if err != nil {
				fmt.Println(err)
				return false
			}

			m[lsl.Metric] = lsl
			//fmt.Println("lsl", lsl)
		}
		/*if l == 8 {
			lsl.SumSquared, err = strconv.ParseFloat(string(as[7]), 64)
			if err != nil {
				fmt.Println(err)
				return false
			}
			m[lsl.Metric]=lsl
			//fmt.Println("lsl8", lsl)
		}*/
	}

	return true
}

