# Lustrebeat

Welcome to Lustrebeat.

Lustrebeat was a brute-force attempt, made in 2017, to grab all the metrics from various "stats" files of the Lustre filesystem, and that without much thinking. So the design decisions were as follows.

* Using Golang glob to check for the stats files and grab whatever found. This means that the beat should run on the host (be that client, MDS, or OSS) it collects metrics from. 
* Tag the values by the components of the path (FS name, OST name, etc.) of each of the stat files.
* Using a simple parser for {llite, OST, MDT} stats files that would be agnostic to the names of the fields there. Every line will be collected.
* Lustre Job Stats: they are actually Yaml. Using Go-Yaml2 library to parse them in one blow, again, collecting every metric there.
* Single ES document per stat file (as opposed to single-metric-value documents).
* Following the Metricbeat ideology, pass the raw metrics/counters to ES or LS, withot any attempts to calculate rates etc. Everything to be done on the ES side.
* In addition to Lustre stats, generic host metrics (CPU, Memory, Network etc.) were added, using the excellent [Shirou gopsutil library](https://github.com/shirou/gopsutil).
* Plus Infiniband counters and ZFS pool and stats since interconnect is often IN or OPA and the storage backend is ZFS.

The drawback of this approach is obvious: there are too many metrics collected on any Lustre system in production. Especially if Exports and Jobstats collection is turned on. On the other hand, if you want to explore what Lustre or ZFS has in its stats files, this beat gives you every field present.  

## Getting Started with Lustrebeat

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/gshamov/lustrebeat`
That is, ensure that GOPATH is set and you have checked out this repository being under ${GOPATH}/src/github.com/gshamov/

### Requirements

* [Golang](https://golang.org/dl/) 1.8 - 1.10
* Python with cookiecutter installed. 

On ComputeCanada systems the above means 

```
module load go; module load python
virtualenv ck; source ck/bin/activate; pip install cookiecutter
```

* Checkout a "good" version of beats into $GOPATH as follows. This one tested with 6.1.x .

```
mkdir -p $GOPATH/src/github.com/elastic/
cd $GOPATH/src/github.com/elastic/
git clone -b v6.1.4 https://github.com/elastic/beats/
```

* The following Golang packages: needs to be go get'd

```
go get github.com/gshamov/gopsutil/mem  # that one isnt really necessary, was an experiment for committed_as. Shirou does it all!
go get github.com/shirou/gopsutil
go get gopkg.in/yaml.v2	
go get github.com/Masterminds/glide
```

### Init Project
To get running with Lustrebeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Lustrebeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/gshamov/lustrebeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Lustrebeat run the command below. This will generate a binary
in the same directory with the name lustrebeat.

```
make
```


### Run

To run Lustrebeat with debugging output enabled, run:

```
./lustrebeat -c lustrebeat.yml -e -d "*"
```

Check lustrebeat.yml for the config options. GAS: as of now the "full" config is old and boilerplate. Not updated yet.


### Test

GAS: the documentation below is a boilerplate thing from Beats library. I doubt the tests work as of now.

To test Lustrebeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```


### Cleanup

To clean  Lustrebeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Lustrebeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/src/github.com/gshamov/lustrebeat
git clone https://github.com/gshamov/lustrebeat ${GOPATH}/src/github.com/gshamov/lustrebeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make package
```

This will fetch and create all images required for the build process. The hole process to finish can take several minutes.

## Missing things, TODO ##

A lot is missing. Fields/templates to be done; tests, documentation, packaging as well. 
To linit the amount of ES data it generates, a filtering mechanism perhaps would help. 

## Similar / better projects

* [HPE Prometheus Lustre Exporter](https://github.com/HewlettPackard/lustre_exporter)
* [Telegraf lustre2 module](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/lustre2)
