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
