// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

	// what to collect. Llite is only on client, Osc on client and mds, Mds and Oss are them.
	// jobstats are separate config item
	// exports are also a separate config item

// hardcoding the Client stats. Assumed FS name 'lustre'
//const llite_stats = "/proc/fs/lustre/llite/lustre-*/stats" // one per FS
//const osc_stats = "/proc/fs/lustre/osc/lustre-OST*/stats"  // one per FS per OST

//Some of MDS stats. No exports yet.
//const mds_stats string = "/proc/fs/lustre/mdt/*/md_stats"
//const mds_jobstat string = "/proc/fs/lustre/mdt/*/job_stats"

//some of OSS stats. No exports yet.

//const oss_stats string = "/proc/fs/lustre/obdfilter/*/stats"
//const oss_jobstat string = "/proc/fs/lustre/obdfilter/*/job_stats"

type Config struct {
	Period time.Duration `config:"period"`
	
	Zfs ZfsConfig `config:"zfs"`

	Host HostConfig `config:"host"`

	Lustre LustreConfig `config:"lustre"`
}

type LustreConfig struct {
	//client ; llite is aggregated, osc gives per ost lov level stats
	Llite     bool `config:"llite"`
        Osc       bool `config:"osc"`
	
	//mds ; note that osc enables some output on mds as well
	Mds       bool `config:"mds"`

	//oss
	Oss       bool `config:"oss"`
	
	// extra outputs for mds, oss have effect on both
	Jobstats bool `config:"jobstats"`
	Exports  bool `config:"exports"`
	
	//paths
	LliteStatsPath string `config:"llite_stats_path"`
	OscStatsPath   string `config:"osc_stats_path"`
	MdsStatsPath   string `config:"mds_stats_path"`
	OssStatsPath   string `config:"oss_stats_path"`
	ExportSuffix   string `config:"exports_suffix"`
	
	// single-numeric stats like kbytesfree, kbytesavail,  collected per ost and mds
	GetNumbers bool `config:"getnumbers"`
	ListOfNumbers []string `config:"list_of_numbers"`

}

type ZfsConfig struct {
	//Zfs stats ans pools
	Stats bool `config:"zfsstats"`
	Pools bool `config:"zfspools"`
	ZfsStatsPath string `config:"zfs_stats_path"`
	ListOfZstats []string `config:"list_of_zstats"`
}

type HostConfig struct {
	// host metrics, from gopsutils and self made for IB
	LoadAvg bool `config:"loadavg"`
	MemStat bool `config:"memstat"`
	IBCounters bool `config:"ibcounters"`
	NetIOCounters bool `config:"netiocounters"`

	IBCountersPath string `config:"ib_counters_path"`

	ListOfCounters []string `config:"list_of_counters"`
	ListOfNetworks []string `config:"list_of_networs"`
	
}

var DefaultConfig = Config{
	Period: 10 * time.Second,
	
	//attempt to make the spagetti yaml hierarchical
	Zfs: DefaultZfsConfig,

	Host: DefaultHostConfig,

	Lustre: DefaultLustreConfig,
}

var DefaultLustreConfig = LustreConfig {
	Llite:     true,
	Osc:       false,
	Mds:       false,
	Jobstats: false,
	Exports:  false,
	Oss:       false,
	
	LliteStatsPath: "/proc/fs/lustre/llite/*", 
	OscStatsPath:   "/proc/fs/lustre/osc/*",  
	MdsStatsPath:   "/proc/fs/lustre/mdt/*",
	OssStatsPath:   "/proc/fs/lustre/obdfilter/*",
	ExportSuffix:   "/exports/*o2ib", // this assumes IB by default, needs changr for tcp or what not

	GetNumbers: false,
	ListOfNumbers:  []string{"filesfree", "filestotal", "kbytesavail", "kbytesfree", "kbytestotal", },

}

var DefaultHostConfig = HostConfig{
	ListOfCounters: []string{"port_rcv_packets", "port_rcv_data", "port_xmit_data", "port_xmit_packets", "symbol_error", },
	ListOfNetworks: []string{"ib0", "lo", },
        LoadAvg: false,
        MemStat: false,
        IBCounters: false,
        NetIOCounters: false,
        // should work also for OPA
        IBCountersPath: "/sys/class/infiniband/*/ports/*/counters",
}

var DefaultZfsConfig = ZfsConfig{
        //raw ZFS stats
	Stats: true,
	Pools: true,       
	ZfsStatsPath:   "/proc/spl/kstat/zfs/*",    
	ListOfZstats:   []string{"dmu_tx", "fm", "arcstats", "xuio_stats", }, // zil and zfetchstats and vdev_cache_stats are all zeroes
}

