package main

import (
	"context"
	"os"
	"flag"
	"fmt"
  "bufio"

	"github.com/docker/docker/client"
  "github.com/docker/docker/api/types"
	/*"github.com/docker/docker/pkg/stdcopy"*/
)

func main() {
	args := os.Args[1:]

  actionPtr := flag.String("action", "build", "Action: build/start/monitor/logs")
  flag.Parse()
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		panic(err)
	}

	cli.NegotiateAPIVersion(ctx)

	if(*actionPtr == "build") {
		fmt.Println("Build from file ", args[2])
		f, err := os.Open(args[2])
		defer f.Close()

		reader := bufio.NewReader(f)
    resp, err := cli.ImageBuild(ctx, reader, types.ImageBuildOptions{})

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(resp)

	}



}
