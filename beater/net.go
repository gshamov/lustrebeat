package beater

import (
	"github.com/shirou/gopsutil/net"
	//"fmt"
)

func GetNetIOCs(valuelist map[string]struct{}) (result []net.IOCountersStat, err error) {
	// the purpose of this function is to filter in selecte network interfaces
	// and may be also to kill dimensionality, making result more IB-like? may be later our outside

	iocs, err := net.IOCounters(true)
	if err != nil {
		//log.Fatal(err)
		return nil, err
	}

	l := len(valuelist)

	if l == 0 {
		result = iocs
	} else {
		result = make([]net.IOCountersStat, 0, l)
		for _, v := range iocs {
			if _, ok := valuelist[v.Name]; ok {
				result = append(result, v)
				//fmt.Println(v)
			}
		}
	}
	return result, err
}
