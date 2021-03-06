# Data Container

First, we can test building a simple image with a go binary entrypoint, [hello.go](hello.go)
as follows:

## Simple Example: Hello World

```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o hello
```
and test running the binary on our machine:

```bash
$ ./hello 
Hello from OS-less container (Go edition)
```

And then add this file to a container with base `scratch` represented in the
[Dockerfile.hello](Dockerfile.hello). For our first effort, we just want to
add the binary:

```
FROM scratch
COPY hello /
CMD ["/hello"]
```

and build.

```bash
docker build -f Dockerfile.hello -t hello .
```

And then test running it!

```bash
$ docker run --rm hello
Hello from OS-less container (Go edition)
```

## Adding Files Only

Now let's try something different - instead of adding an executable, let's
add a folder of files, represented in [Dockerfile.data](Dockerfile.data).

```
FROM scratch
WORKDIR /data
COPY data/*
```

and then build:

```
docker build -f Dockerfile.data -t hello .
```

We quickly learn that we actually need the container to have an entrypoint,
and one that will keep the container running. Let's try doing that in
[sleep.go](sleep.go).

```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o sleep -i sleep.go
```

And then test locally to make sure we sleep:

```bash
$ ./sleep 
Running sleep command...

```
and then update the Dockerfile to run the command, and rebring the image up:

```bash
$ docker build -f Dockerfile.data -t hello .
$ docker-compose up -d
Creating docker_data_1 ... done
Creating docker_base_1 ... done
```

Check the status!

```bash
docker-compose ps
    Name             Command        State   Ports
-------------------------------------------------
docker_base_1   tail -f /dev/null   Up           
docker_data_1   /sleep              Up  
```

Now is the test - we want to be able to shell into the busybox container,
and then see the data mounted from the data container. The shared volume should
be at `/data`.

```bash
$ docker exec -it docker_base_1 sh
/ # ls /data/
avocado.txt  tomato.txt
```

We have data! And it's bound from the otherwise empty container. The docker-compose.yml
looks like this:

```
version: "3"
services:
  base:
    restart: always
    image: busybox
    entrypoint: ["tail", "-f", "/dev/null"]
    volumes:
      - data-volume:/data

  data:
    restart: always
    image: hello
    volumes:
      - data-volume:/data

volumes:
  data-volume:
```

## Testing the Scientific Filesystem

We next want to figure out how to provide functions to query (or otherwise interact)
with the data in the container. I thought it would be fun to first try
adding a scif binary to the container, and we can do this by grabbing
it from [quay.io/scif/scif-go](https://quay.io/repository/scif/scif-go?tab=tags).
Note that it's a multistage build - we watch to just grab the scif
binary and then add it to scratch.

```
FROM quay.io/scif/scif-go:0.0.1.rc as base
FROM scratch
WORKDIR /data
COPY --from=base /usr/local/bin/scif /scif
CMD ["/scif"]
```

Now we can build

```bash
docker build -f Dockerfile.scif -t hello .
```

and test running it - do we hit (and successfully run) the scif binary?

```bash
docker run -it hello
standard_init_linux.go:211: exec user process caused "no such file or directory"
```

That didn't work, so likely we need to interact with scif before we do a multistage
build. The scientific filesystem can install data, and then remove itself. Take
a look at the [recipe.scif](recipe.scif) file for how we create files and data.
We can then update the docker-compose.yml.

```
version: "3"
services:
  base:
    restart: always
    image: busybox
    entrypoint: ["tail", "-f", "/dev/null"]
    volumes:
      - data-volume:/data

  data:
    restart: always
    image: hello
    volumes:
      - data-volume:/data

volumes:
  data-volume:
```

and then remove the previous data volume, and run the orchestration again:

```bsah
docker-compose up -d
$ docker exec -it docker_base_1 sh
```

And we can see the scif data hierarchy at the base of the container!
```bash
/ # ls /scif/
apps  data
/ # ls /scif/data/
hello-custom        hello-world-echo    hello-world-env     hello-world-script
```
Although the dependencies for the associated apps software (in the apps folder)
might not be included in the container, this would still be a way to package
scripts alongside the data. Of course now we would want to figure out how to
package the data, generate a manifest for it, and then query.

## Testing A Memory Database

One thing I want to try is embedding a database into the binary that includes
an easy way to interact with (e.g., search or otherwise query) the data.
My first idea is to use an in memory database, so I'd basically want to:

 1. Embed the SQL schema and rows in the go binary as strings.
 2. Open a new memory database when on init (sql.Open("sqlite3",:memory:`)
 3. Create the schema and insert the rows.

Although there isn't disk access, if we do something simple like:

```go
package main

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

func main() {

    // Open an in memory database
    fmt.Println("Attempting to open in-memory database.")
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(db)
}
```

```bash
go get github.com/mattn/go-sqlite3
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o db -i db.go
$ docker build -f Dockerfile.db -t db .
$ docker run db
```

We get an error:

```
standard_init_linux.go:211: exec user process caused "no such file or directory"
```
Meaning that the sqlite library we are using is likely needing to interact with
the host in some way.

### A Simpler Approach

1. Find an in-memory database
2. Test setting / getting values
3. Create custom container library

Let's try this again, this time with the approach above. This seems to work!

```bash
go get github.com/vsoch/containerdb
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o db -i db.go
```

...build it (we don't need docker-compose since it has an entrypoint)

```bash
$ docker build -f Dockerfile.db -t db .
```

We are able to set and print a value, no OS required.

```bash
$ docker run db
value is myvalue
```

Woot! So next I'm going to, instead of having a single random value added,
have it be so that an actual dataset metadata is added here, and then we expose
functions to interact / query it.  More to come!

## Using cdb

Between the last example and this one, I've developed a Python helper, cdb (container database)
at [vsoch/cdb](https://github.com/vsoch/cdb) that is optimized to extract metadata for files.
Don't worry about that too much for now, but I'll say the basic idea is that it generates
a golang script that is akin to the previous db.go, and then that can be built and added
to a container with scratch. For the real process I'll use multiple multistage builds
and not need to do anything on the host, but for now I'll just follow the same practice
as above. The idea now is that we are working on a template for cdb that will
also add a command line parser to the entrypoint. This work will be in [entrypoint.go].
I'll also want to write functions for basic searching, and indexing.

```
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o entrypoint -i entrypoint.go
$ docker build -f Dockerfile.entrypoint -t entrypoint .
```

We then have a simple way to do the following:

**metadata**

If we just run the container, we get a listing of all metadata alongside the key.

```bash
$ docker run entrypoint 
/data/avocado.txt {"size": 9, "sha256": "327bf8231c9572ecdfdc53473319699e7b8e6a98adf0f383ff6be5b46094aba4"}
/data/tomato.txt {"size": 8, "sha256": "3b7721618a86990a3a90f9fa5744d15812954fba6bb21ebf5b5b66ad78cf5816"}
```

We can also just list data files with `-ls`

```bash
$ docker run entrypoint -ls
/data/avocado.txt
/data/tomato.txt
```

Or we can list ordered by one of the metadata items:

```bash
$ docker run entrypoint -metric size
Order by size
/data/tomato.txt: {"size": 8, "sha256": "3b7721618a86990a3a90f9fa5744d15812954fba6bb21ebf5b5b66ad78cf5816"}
/data/avocado.txt: {"size": 9, "sha256": "327bf8231c9572ecdfdc53473319699e7b8e6a98adf0f383ff6be5b46094aba4"}
```

Or search for a specific metric based on value.

```bash
$ docker run entrypoint -metric size -search 8
/data/tomato.txt 8

$ docker run entrypoint -metric sha256 -search 8
/data/avocado.txt 327bf8231c9572ecdfdc53473319699e7b8e6a98adf0f383ff6be5b46094aba4
/data/tomato.txt 3b7721618a86990a3a90f9fa5744d15812954fba6bb21ebf5b5b66ad78cf5816
```

Or we can get a particular file metadata by it's name:

```bash
$ docker run entrypoint -get /data/avocado.txt
/data/avocado.txt {"size": 9, "sha256": "327bf8231c9572ecdfdc53473319699e7b8e6a98adf0f383ff6be5b46094aba4"}
```

or a partial match:

```bash
$ docker run entrypoint -get /data/
/data/avocado.txt {"size": 9, "sha256": "327bf8231c9572ecdfdc53473319699e7b8e6a98adf0f383ff6be5b46094aba4"}
/data/tomato.txt {"size": 8, "sha256": "3b7721618a86990a3a90f9fa5744d15812954fba6bb21ebf5b5b66ad78cf5816"}
```

Okay, so I think this is a good start for a generic filestructure of data! I'll update
the cdb library to use this template.
