package main

import (
	"context"
	"os"
	"fmt"
	"io"
	"bufio"
	"strings"
	"errors"
	"github.com/tidwall/gjson"
  "github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/api/types/container"
)

func CheckErr(e error) {
    if (e != nil) {
        fmt.Println("Error:", e.Error())
				os.Exit(1)
    }
}

func GetContext(path string) io.Reader {
     ctx,_ := archive.TarWithOptions(path, &archive.TarOptions{})

     return ctx
}

func StoreToFile(filename string, text string) error{
     file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
     defer file.Close()

     if err != nil {
	    return err
     }

     file.WriteString(text+"\n")

		 return nil
}

func ReadFromFile(filename string) ([]string, error) {
    var data []string
    file, err := os.Open(filename)
		defer file.Close()

		if err != nil {
			return nil, err
		}

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
			data = append(data, scanner.Text())
    }

    return data, nil
}

func Build(client *client.Client, ctx context.Context, path string) (string,error) {
	resp, err := client.ImageBuild(ctx, GetContext(path), types.ImageBuildOptions{})

	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
			index := strings.Index(scanner.Text(), "Successfully built")
			if index > 1 {
				 imageId := scanner.Text()[index+19:len(scanner.Text())-4]
				 return imageId, nil
			}
	}

	return "", errors.New("Build failed")
}

func Run(client *client.Client, ctx context.Context, imageId string) (string, error) {
	resp, err := client.ContainerCreate(ctx, &container.Config{Image: imageId}, nil, nil, "")

	if err != nil {
		 return "", err
	}

	if client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		 return "", err
	}

	return resp.ID, nil
}

func Validate(client *client.Client, ctx context.Context, containersCreated []string) error{
	containersRunning, err := client.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return err
	}

	for _, c := range containersCreated {
		 isRunning := false
		 for _, r := range containersRunning {
			 if c == r.ID {
				 isRunning = true
				 fmt.Println("Container ", c, "is running")
				 break
			 }
		 }

		 if !isRunning {
			 fmt.Println("Container ", c, "is not running")
		 }
	}

	return nil
}

func Monitor(client *client.Client, ctx context.Context, containersCreated []string) error {
	for _, c := range containersCreated {
		fmt.Println("Container stats:", c)
		resp, err := client.ContainerStats(context.Background(), c, false)

		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Scan()
		bytes := scanner.Bytes()
		mem_usage := gjson.GetBytes(bytes, "memory_stats.usage")
		cpu_usage := gjson.GetBytes(bytes, "cpu_stats.cpu_usage.total_usage")
		rx_bytes := gjson.GetBytes(bytes, "networks.eth0.rx_bytes")
		tx_bytes := gjson.GetBytes(bytes, "networks.eth0.tx_bytes")

		fmt.Println("Memory usage:", mem_usage, "CPU usage:", cpu_usage, " I/O:", rx_bytes, "/", tx_bytes)
 }

 return nil
}

func Logs(client *client.Client, ctx context.Context, containersCreated []string) (string,error) {
	logs := ""
	for _, c := range containersCreated {
		reader, err := client.ContainerLogs(ctx, c, types.ContainerLogsOptions{ShowStdout: true})
    if err != nil {
			return "", err
		}
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			var s = []string {logs, c+":"+scanner.Text()}
			logs = strings.Join(s, "\n")
		}

	}
	return logs,nil
}

func Clean(client *client.Client, ctx context.Context, containersCreated[] string, imageId string) error {
	for _, c := range containersCreated {
		err := client.ContainerStop(ctx, c, nil)
		if err != nil {
			return err
		}

		err = client.ContainerRemove(ctx, c, types.ContainerRemoveOptions{})

		if err != nil {
			return err
		}
	}

	_, err := client.ImageRemove(ctx, imageId, types.ImageRemoveOptions{})

	if err != nil {
		return err
	}

	return nil
}

func main() {
	args := os.Args[1:]
	ctx := context.Background()

  if len(args) < 1 {
	   fmt.Println("Usage: omega <action>. Action: build, run, validate, monitor, logs")
		 os.Exit(0)
	}

	client, err := client.NewClientWithOpts(client.FromEnv)
	CheckErr(err)

	client.NegotiateAPIVersion(ctx)
  action := args[0]

	if(action == "build") {
		  if len(args) == 1 {
				fmt.Println("Path is missing")
				os.Exit(1)
			}
     	fmt.Println("Build from directory:", args[1])
			imageId,err := Build(client, ctx, args[1])
			CheckErr(err)
			err = StoreToFile("/tmp/imageid", imageId)
			CheckErr(err)
			fmt.Println("A new container image is created with id:", imageId)

	}else if action == "run" {
	  if data, err := ReadFromFile("/tmp/imageid"); (err == nil && len(data) > 0) {
			containerId, err := Run(client, ctx, data[0])
			CheckErr(err)
			StoreToFile("/tmp/containers", containerId)
			fmt.Println("A new container is running with id:", containerId)
		}else {
			CheckErr(err)
			fmt.Println("No created image")
		}
	}else if action == "validate" {

		if containers, err := ReadFromFile("/tmp/containers");(err == nil && len(containers) > 0) {
			err = Validate(client, ctx, containers)

			CheckErr(err)
		}else {
			CheckErr(err)
			fmt.Println("No running containers")
		}
	}else if action == "monitor" {
     if containers, err := ReadFromFile("/tmp/containers"); (err == nil && len(containers) > 0) {
			 err = Monitor(client, ctx, containers)

			 CheckErr(err)
		 }else {
			 CheckErr(err)
			 fmt.Println("No running containers")
		 }

	 }else if action == "logs" {

		 if containers, err := ReadFromFile("/tmp/containers"); (err == nil && len(containers) > 0) {
			 logs,err := Logs(client, ctx, containers)
			 CheckErr(err)

			 fmt.Println(logs)
		 }else {
			 CheckErr(err)
			 fmt.Println("No running containers")
		 }

 	}else if action == "clean" {
		 data, err := ReadFromFile("/tmp/imageid")
		 CheckErr(err)

		 if len(data) == 0 {
			 fmt.Println("No image")
			 os.Exit(1)
		 }

     if containers, err := ReadFromFile("/tmp/containers"); (err == nil && len(containers) > 0) {
			 err := Clean(client, ctx, containers, data[0])
			 CheckErr(err)
			 os.Remove("/tmp/containers")
			 os.Remove("/tmp/imageid")
		 }else {
			 CheckErr(err)
			 fmt.Println("No running containers")
		 }
	}else {
		fmt.Println("Invalid argument")
	}
}
