package main

import (
	"context"
	"os"
	"fmt"
        "io"
        "bufio"
        "strings"

	"github.com/docker/docker/client"
        "github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/api/types/container"
)

func GetContext(path string) io.Reader {
     ctx,_ := archive.TarWithOptions(path, &archive.TarOptions{})

     return ctx
}

func StoreToFile(filename string, text string) { 
     file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)	
     defer file.Close()

     if err != nil {
	    fmt.Println("Could not write to file", err)
	    return 
     }

     file.WriteString(text+"\n")
}

func ReadFromFile(filename string) []string {
    var data []string
    file, _ := os.Open(filename)
    
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
	data = append(data, scanner.Text())
    }

    return data
}

func main() {
	args := os.Args[1:]

	action := args[0]
        fmt.Println(action)
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	cli.NegotiateAPIVersion(ctx)

	if(action == "build") {
		fmt.Println("Build from directory:", args[1])

                resp, err := cli.ImageBuild(ctx, GetContext(args[1]), types.ImageBuildOptions{})
                
		if err != nil {
		  fmt.Println(err)
		}

		scanner := bufio.NewScanner(resp.Body)
                
		for scanner.Scan() {
                    index := strings.Index(scanner.Text(), "Successfully built") 
		    if index > 1 {
		       imageId := scanner.Text()[index+19:len(scanner.Text())-4]
		       fmt.Println("The image is created with id:", imageId)
		       StoreToFile("/tmp/imageid", imageId)
		    }
		}

	}else if action == "run" {
	 fmt.Println("Run containers", args[1])
	 data := ReadFromFile("/tmp/imageid")

	 resp, err := cli.ContainerCreate(ctx, &container.Config{Image: data[0]}, nil, nil, "") 

	 if err != nil {
           panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
	   panic(err)
	}

	fmt.Println("A new container is running with:", resp.ID)
      }
}
