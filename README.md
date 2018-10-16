# Lustrebeat

Welcome to Lustrebeat.

Lustrebeat was a brute-force attempt, made in 2017, to grab all the metrics from various "stats" files of the Lustre filesystem, and that without much thinking. So the decisions as follows.

* Using Golang glob to check for the stats files and grab whatever found. Tag the values by the components of the path (FS name, OST name, etc.)
* Using a simple parser for stats files that would be agnostic to the names of the counters. Every line will be collected.
* Following the Metricbeat ideology, pass the raw counters to ES or LS, withot any attempts to calculate rates etc. Everything to be done on the ES side.
* In addition to Lustre stats, generic host metrics (CPU, Memory, Network etc.) were added, using the excellent shirou library.
* Plus Infiniband counters and ZFS pool and stats


## Getting Started with Lustrebeat

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/gshamov/lustrebeat`

### Requirements

* [Golang](https://golang.org/dl/) 1.8 - 1.10
* Python with cookiecutter installed. 

On ComputeCanada systems the above means 

```
module load go; module load python
virtualenv ck; source ck/bin/activate; pip install cookiecutter
```

* The following Golang packages: needs to be go get'd

```
go get github.com/gshamov/gopsutil/mem
go get github.com/shirou/gopsutil
go get gopkg.in/yaml.v2	
```

* And, finally, checkout a "good" version of beats into $GOPATH as follows. This one tested with 6.1.x .

```
mkdir -p $GOPATH/src/github.com/elastic/
cd $GOPATH/src/github.com/elastic/
git clone -b v6.1.4 https://github.com/elastic/beats/
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
