package beater

import (
	"bytes"
	//"fmt"
	"io/ioutil"
	"log" // how do I do logging from beats? May be their logp?
	"path/filepath"
	"strconv"
)

// performance counters for infiniband and single-value metrics for lustre
// are files with single counter in them. 

func GetNumberVal(s string) (v uint64, ok bool) {
	// opens a file that contains single numerical value,
	// like /proc/fs/lustre/llite/lustrefs-ffff880152937000/kbytestotal
	// or an infiniband counters/port_xmit_data
	buf, err := ioutil.ReadFile(s)
	v = 0
	ok = true
	if err != nil {
		//log.Fatal(err)
		ok = false
	}
	// add a regex to check if the value is single int number ? (can be a float?)
	// for now just trusting it to be a right file
	v, err = strconv.ParseUint(string(bytes.TrimSpace(buf)), 10, 64) // should it be float?
	if err != nil {
		//log.Fatal(err)
		ok = false
	}
	// silently return false if something fails for now!
	return v, ok
}

func GetValFiles(statsdir string, valuelist map[string]struct{}) (result map[string]uint64, err error) {
	// gets a list of values from under path by files named like []valuelist.
	// the path can have stars for the hash, but better be under same OST/MFS/FIlesystem?
	// or no stars at all? Let it be no stars for simplicity

	filelist, err := filepath.Glob(statsdir + "/*")
	if err != nil {
		log.Fatal(err)
	}
	
	result = make(map[string]uint64, len(filelist))
	
	//fmt.Println("from getValFiles", statsdir, valuelist, filelist)
		
	all :=  ( len(valuelist) == 0 )
	//fmt.Println("all", all)
	
	for _, s := range filelist {
		cs := filepath.Base(s)
		var ok bool = false
		
		if ! all {
		    _, ok = valuelist[cs]
		}
		
		if ok || all {
			if v, ok1 := GetNumberVal(s); ok1 {
				result[cs] = v
				//fmt.Println(result, v)
			}
			// else {
			//} Do I care about else? Will be no value
		}
	}
	return result, err

}


func GetIBCounters(path string, valuelist map[string]struct{}) (result map[string]uint64, err error) {
	//result = GetAllValFiles(path)
	//empty := map[string]struct{}{}
	//notempty := map[string]struct{} {"port_rcv_packets":{}, "port_rcv_data":{}, "port_xmit_data":{}, "port_xmit_packets":{}, "symbol_error":{},}

	result, err = GetValFiles(path, valuelist)

	// some ot the Mellanox port counters are 4x because of 4 lanes?
	// https://community.mellanox.com/docs/DOC-2751 https://community.mellanox.com/docs/DOC-2572

	singlenames := map[string]struct{}{"port_rcv_data": {}, "port_xmit_data": {}} 
	// the names to be separately handled
	
	for k, v := range result {
			if _, ok2 := singlenames[k]; ok2 {
				result[(k + "4")] = v / 4
			}
	}

	return result, err
}

