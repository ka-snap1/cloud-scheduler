package main

import (
	"fmt"
	"log"

	"cloud-scheduler/internal/cloud/aliyun"
)

func main() {
	client, err := aliyun.CreateClient()
	if err != nil {
		log.Fatalf("create client failed: %v", err)
	}

	instanceId, err := aliyun.CreateInstance(client)
	if err != nil {
		log.Fatalf("create instance failed: %v", err)
	}

	if err := aliyun.StartInstance(client, instanceId); err != nil {
		log.Fatalf("start instance failed: %v", err)
	}

	fmt.Printf("instance created and started: %s\n", instanceId)
}
