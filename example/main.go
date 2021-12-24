package main

import (
	"fmt"
	"goconsul"

	"github.com/hashicorp/consul/api"
)

func testKV(client *goconsul.Client) {

	cmd := goconsul.NewKVCmd(client, "")
	if err := cmd.SetStr("key", "value"); err != nil {
		panic(err)
	}

	var val string
	val, err := cmd.GetStr("key")
	if err != nil {
		panic(err)
	}

	fmt.Println("get val: ", val)

	if err := cmd.Delete("key"); err != nil {
		panic(err)
	}
}

func testDiscovery(client *goconsul.Client) {
	ins := goconsul.NewInstance("test-service",
		"127.0.0.1",
		80,
		nil,
		nil,
		"instance-1",
		nil,
		nil,
		client,
		"test",
	)
	ins.Check = &goconsul.ServiceCheck{
		AgentServiceCheck: api.AgentServiceCheck{
			HTTP:     "http://",
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	if err := ins.Register(); err != nil {
		panic(err)
	}

	insArray, err := client.DiscoverInstancesWithName("test-service", nil, 1)
	if err != nil {
		panic(err)
	}

	fmt.Println("discover services:", *insArray[0])

	ins.Deregister()
}

func main() {
	client := &goconsul.Client{}
	if err := client.Connect("127.0.0.1", 8500, ""); err != nil {
		panic(err)
	}

	testKV(client)
	testDiscovery(client)
}
