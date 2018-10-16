package beater

import (
	//"bytes"
	"fmt"
	"log" // how do I do logging from beats? May be their logp?
	"path/filepath"

	"strings"
	"time"

	"github.com/gshamov/gopsutil/mem"
	//"github.com/shirou/gopsutil/cpu"
	//"github.com/shirou/gopsutil/disk"
	//"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	//"github.com/shirou/gopsutil/net"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/gshamov/lustrebeat/config"
)

// boilerplate
type Lustrebeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Lustrebeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

// boilerplate function
func (bt *Lustrebeat) Run(b *beat.Beat) error {
	logp.Info("Lustrebeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period) // leftover from the boilerplate,
	//I use it to see if all works with config disabling everything else
	var counter uint64 // changed the type to coulnt carelessly!
	counter = 1

	// actually the counter it can be of use if we want to (un)skip every n-th beat
	// for this or that metric

	// Should learn how to use the config. Ok, it actually is bt.config.SomeName
	// Defaults supposed to come from config.go so I do no checking here.
	// Should these guys go to some init() or create() function?
	
        fmt.Println("Config::", bt.config)

	llite := bt.config.Lustre.Llite
	osc := bt.config.Lustre.Osc
	oss := bt.config.Lustre.Oss
	mds := bt.config.Lustre.Mds

	
	// exports are on for each of OSS, MDS. 
	// Should I have selected jobstats as well as a single parameter?
	exports := bt.config.Lustre.Exports
	jobstats := bt.config.Lustre.Jobstats
	
	// we can also collect total KB counts from OSS, MDS stats.
	getnum := bt.config.Lustre.GetNumbers
	listnum := List2Set(bt.config.Lustre.ListOfNumbers)

	// infiniband and network metrics now can filter for specific things too, more lists

	listcounters := List2Set(bt.config.Host.ListOfCounters) // empty lists should work too, test that!
	listnetworks := List2Set(bt.config.Host.ListOfNetworks)

	// lets try if lists work from config. Comment in production.
	//fmt.Printf("%T %v\n", bt.config, bt.config)

	//Zfs stats if requested
	zss := bt.config.Zfs.Stats
	zps := bt.config.Zfs.Pools
	listzstats := List2Set(bt.config.Zfs.ListOfZstats)
	fmt.Printf("%T %v\n", bt.config.Zfs, bt.config.Zfs)

	// glob the ZFS paths outside of the main loop
	var ioz, zstats []string

	if zps {
		// is Glob again any better?
		ioz, err = filepath.Glob("/proc/spl/kstat/zfs/*/io")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%T %v\n", ioz, ioz)
	}

	if zss {
		zstats, err = filepath.Glob(bt.config.Zfs.ZfsStatsPath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%T %v\n", zstats, zstats)
	}

	// ok, lets collect all the stats that are the easy stats. Local FS is assumed, and io.ReadFile used
	// the existencial assumptions is that Glob sees the files on a local FS.
	// A refactoring to remote file access or lctl commands might need more checking.

	// TODO: extract FS name and OSS name from procs filenames and tag records properly.

	statslist := []string{}          // generic object-wide stats
	expstats := []string{}           // lets have a separate list for exports for there are many of them
	kbstats := make(map[string]bool) // a set of places with things like kbtotal

	if llite {
		llts := bt.config.Lustre.LliteStatsPath + "/stats"
		statslist0, err := filepath.Glob(llts)
		if err != nil {
			log.Fatal(err)
		}
		statslist = append(statslist, statslist0...)
	}

	if osc {
		oscs := bt.config.Lustre.OscStatsPath + "/stats"
		statslist0, err := filepath.Glob(oscs)
		if err != nil {
			log.Fatal(err)
		}
		statslist = append(statslist, statslist0...)
		if getnum {
			for _, s := range statslist0 {
				kbstats[s] = true
			}
		}
	}

	if mds {
		mdss := bt.config.Lustre.MdsStatsPath + "/md_stats" // note that MDS-wide is not stats but md_stats
		statslist0, err := filepath.Glob(mdss)
		if err != nil {
			log.Fatal(err)
		}
		statslist = append(statslist, statslist0...)
		if getnum {
			for _, s := range statslist0 {
				kbstats[s] = true
			}
		}
		if exports {
			mdex := bt.config.Lustre.MdsStatsPath + bt.config.Lustre.ExportSuffix + "/stats"
			statslist1, err := filepath.Glob(mdex)
			if err != nil {
				log.Fatal(err)
			}

			expstats = append(expstats, statslist1...)
		}
	}

	if oss {
		osss := bt.config.Lustre.OssStatsPath + "/stats"
		statslist0, err := filepath.Glob(osss)
		if err != nil {
			log.Fatal(err)
		}
		statslist = append(statslist, statslist0...)
		if getnum {
			fmt.Println(statslist0)

			for _, s := range statslist0 {
				kbstats[s] = true
			}
		}
		if exports {
			oex := bt.config.Lustre.OssStatsPath + bt.config.Lustre.ExportSuffix + "/stats"
			statslist1, err := filepath.Glob(oex)
			if err != nil {
				log.Fatal(err)
			}

			expstats = append(expstats, statslist1...)
		}
	}

	fmt.Println(statslist, getnum, "kbstats::: ", kbstats)

	// Jobstats probably need another list for them

	joblist := []string{}
	// re := regexp.MustCompile(`([a-z]+\:)`) // regexp for jobstats max:12345 out of the loop; outsourced

	if jobstats && mds {
		jms := bt.config.Lustre.MdsStatsPath + "/job_stats"
		jm0, err := filepath.Glob(jms)
		if err != nil {
			log.Fatal(err)
		}
		joblist = append(joblist, jm0...)
	}

	if jobstats && oss {
		jos := bt.config.Lustre.OssStatsPath + "/job_stats"
		jo0, err := filepath.Glob(jos)
		if err != nil {
			log.Fatal(err)
		}
		joblist = append(joblist, jo0...)
	}

	// would be good to add a check above? llite is only on the clients.
	// unlikely that MDS and OSS sts are available on the same machine either.

	// MAIN LOOP from the Beat boilerplate
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		if len(statslist) > 0 { // redundant because for?
			for _, statfile := range statslist {
				result, err := GetStatsFile1(statfile)
				//GAS debug
				//fmt.Println(result,err,"LUSTRE STAT DEBUG")
				//GAS debug
				if err != nil {
					continue
				}

				// at this point I also need to know which stat is which?
				// to determine if I want kbytes. ToFIx: doesnt work, doesnt publish either
				if getnum {
					if _, ok := kbstats[statfile]; ok {
						fmt.Println("KBSTATS::::::", statfile, listnum)
						// got either MDS or OST stats file here; get dirname and extract values
						statsdir := filepath.Dir(statfile)
						kbs, _ := GetValFiles(statsdir, listnum)
						fmt.Println("KBS:: ", kbs)
						for k, v := range kbs {
							// probably need to add check if k already in
							result[k] = v
						}
					}
				}
				//fmt.Println(result, err)
				fs, obj, _, _ := MakeLustreTagsE(filepath.Dir(statfile)) // somewhat gross to parse-unparse it all the time
				event_llite := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":   "lustrestat",
						"stats":  result,
						"fs":     fs,
						"object": obj,
					},
				}
				bt.client.Publish(event_llite)
				logp.Info("Event_llite sent")
			}
		}
		// llite and general stats done

		// lets do exports if any; nasty code duplication as usual
		if len(expstats) > 0 { // redundant because for?
			for _, statfile := range expstats {
				result, err := GetStatsFile1(statfile)
				if err != nil {
					continue
				}

				//fmt.Println("Export", result, err)
				fs, obj, ex, _ := MakeLustreTagsE(filepath.Dir(statfile)) // somewhat gross to parse-unparse it all the time

				event_llite := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":   "lustreexport", // should it be same? idk
						"stats":  result,
						"fs":     fs,
						"object": obj,
						"export": ex,
					},
				}
				bt.client.Publish(event_llite)
				logp.Info("Event_llite sent")
			}
		}
		// exports done

		// jobstats loop begins
		if len(joblist) > 0 {
			//fmt.Println(joblist)
			for _, s := range joblist {
				// outsourced the spagetti into jobstat.go
				err := JobStat(bt, s)

				if err != nil {
					fmt.Println(err)
					// what to do? to stop or not to stop? Logp?
				}
			}
			// jobstats loop end
		}

		// generic load metrics (mostly) from gopsutil
		// This is down to Host. config structure now
		//LoadAvg: false,
		//MemStat: false,
		//IBCounters: false,
		//NetIOCounters: false,
		//fmt.Println("config.Host.LoadAvg: ", bt.config.Host.LoadAvg, "config.Host.MemStat: ", bt.config.Host.MemStat,)
		
		if bt.config.Host.LoadAvg || bt.config.Host.MemStat { // why bother separating them, either returns both
			av, _ := load.Avg()           // can it ever fail?
			mem, _ := mem.VirtualMemory() // can it ever fail?
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"type":    "host",
					"loadavg": av,
					"mem":     mem,
				},
			}
			bt.client.Publish(event)
			logp.Info("Event sent")
		}

		// net devices counters
		if bt.config.Host.NetIOCounters {
			iocs, err := GetNetIOCs(listnetworks)
			if err != nil {
				log.Fatal(err) // need to handle it somehow. Just skipping the beat?
			}
			// iocs is an array, convert to separate events to make it easier for Kibana
			for _, v := range iocs {
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":  "network",
						"netio": v,
					},
				}
				bt.client.Publish(event)
				logp.Info("Event sent")
			}
		}

		// Infiniband port counters
		if bt.config.Host.IBCounters {
			// added more logic to traverse/tag multiple NICs and ports
			// lets unglob the paths only
			fl, err := filepath.Glob(bt.config.Host.IBCountersPath)
			if err != nil {
				log.Fatal(err)
				// no IB? to check what happens if there is no preceding path
			}
			for _, s := range fl {
				zz := strings.Split(s, "/")
				//fmt.Println(zz, zz[4], zz[6])
				nic := zz[4]
				port := zz[6]
				ibcs, _ := GetIBCounters(s, listcounters) // remember only give path no added star
				// I have no error handling yet. Do I need it? Should two globs handle it?
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type": "ibcounters",
						"ib":   ibcs,
						"port": port,
						"nic":  nic,
						//"path": s, // do we need to store the path in production?
					},
				}
				bt.client.Publish(event)
				logp.Info("Event sent")
			}
		}

		if zps {
			// ioz defined above, it is the list of pool/io paths
			for _, v := range ioz {
				// make the tag
				pool := strings.Split(v, "/")[5]
				ios, err := GetZfsPoolIofile(v)
				if err != nil {
					log.Fatal(err) // is in fatal here?
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
			}
		}

		if zss {

			for _, s := range zstats {

				cs := filepath.Base(s)
				var ok bool = false
				// no provision for empty lists! should be explicit stats always
				// or nothing will be returned
				
				if _, ok = listzstats[cs]; ok {
				   result, err := GetZfsStatfile(s)
				    //fmt.Println(result, err, s)
				    if err != nil {
					log.Fatal(err)
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
			}
		}

		// the original counter
		// remve from prod version?
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		// // trying to outsource it. Works if beat bt passed as follows.
		//Counter(bt, "counter2", counter)
		counter++
		if counter > 2147483647 {
			counter = 1
		}

	}
}

func (bt *Lustrebeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
