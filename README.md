# Omega tool

## Description
A command line tool to manage containers. The omega builds, runs, validates and monitors a list of containers.

## Implementation
I used the Go language to develop the tool.
It takes as argument a directory with Dockerfile and app. It builds the image and stores the id in the file /tmp/imageid.
Each time a new container is going to be started from omega, the container's id is added in the file /tmp/containers.
In order to execute the other actions, the omega tool reads the list of the containers from the file /tmp/containers
and runs the function which applies the action.

## Installation
You have to install two Go libraries.
The library for the Docker SDK and a library for JSON handling.

Run the commands:
```
go get github.com/docker/docker/client
go get -u github.com/tidwall/gjson
```

In order to build the omega, run the command:
```
make
```

or

```
go build -o omega src/main.go
```

##Arguments

### build
It takes a directory with docker files as argument. It must contains the Dockerfile

Example:
```
omega build <directory>
```

### run
It creates a new container instance and it prints out the id.

Example:
```
omega run
```

### validate
It validates if the created containers are running or not.

Example:
```
omega validate
```

### monitor
It prints out information about the memory, cpu and network input-output usage for each container.

Example:
```
omega monitor
```

### logs
It collects the stdout output of each container and prints out the collected output.

Example:
```
omega logs
```

### clean
It cleans the resources. It stops and deletes the containers. Eventually it removes the image.

Example:
```
omega clean
```
