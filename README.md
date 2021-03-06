# Data Containers

Back in 2016, there was discussion and excitement for data containers.

Two recent developments have told me that now is the time to address this once
more:

 1. An [article](https://iximiuz.com/en/posts/not-every-container-has-an-operating-system-inside) that details this idea, that containers don't necessary need an operating system.
 2. The ability to create a container from scratch supported by Singularity (pull request [here](https://github.com/hpcng/singularity-userdocs/pull/328))

## Needs of a Data Container

Before we can build a data container, we need to decide what is important for it
to have, or generally be. If we think of a "normal" container as providing a base
operating system to support libraries, small data files, and ultimately running
scientific software, then we might define a data container as:

1. a container to support the provenance, management, and query of data.
2. container should work bound as a volume or on it's own
3. it should be possible to customize how metadata is extracted from the dataset

## Development

See the [devel](devel) folder for early development work. 

## Data Containers

This repository will soon be populated with specific docker and singularity examples, generated
by way of [vsoch/cdb](https://github.com/vsoch/cdb), the container database
metadata generator.
