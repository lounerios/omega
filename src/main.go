package main

import (
	"context"
	"os"
	"fmt"
	"io"
	"bufio"
	"strings"
	"encoding/json"
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
     file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
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
		fmt.Println("Start")
		data := ReadFromFile("/tmp/imageid")
		resp, err := cli.ContainerCreate(ctx, &container.Config{Image: data[0]}, nil, nil, "")

    if err != nil {
       panic(err)
	  }

	  if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
	     panic(err)
	  }

	  fmt.Println("A new container is running with:", resp.ID)
		StoreToFile("/tmp/containers", resp.ID)
   }else if action == "validate" {

      buildContainers := ReadFromFile("/tmp/containers")

		 containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})

		 if err != nil {
			 panic(err)
		 }

		 for _, container := range containers {
		    fmt.Println(container.ID)

				for _, c := range buildContainers {

					if container.ID == c {
						fmt.Println("Container is running", container.ID)
					}
				}
	}



	 }else if action == "monitor" {
     buildContainers := ReadFromFile("/tmp/containers")

		 for _, c := range buildContainers {
			 fmt.Println("Container:", c)
			 resp, err := cli.ContainerStats(context.Background(), c, false)

			 if err != nil {
				 panic(err)
			 }

			 dec := json.NewDecoder(resp.Body)
			 var s map[string]interface{}
			if err := dec.Decode(&s); err != nil {
					 fmt.Println(err)
					 return
			}

      cpu_stats := s["cpu_stats"]
      fmt.Println(cpu_stats)

			fmt.Println(s["memory_stats"])
    }
	 }else if action == "logs" {
		 fmt.Println("logs")

 	}else if action == "drop" {
     fmt.Println("drop")
	}
}
