package main

import (
	"fmt"
	"log"
	"time"

	"cloud-scheduler/internal/cloud/aliyun"
	//util "github.com/alibabacloud-go/tea-utils/v2/service"
)

func main() {
	// 创建阿里云ECS客户端
	client, err := aliyun.CreateClient()
	// 预检请求
	// true：发送检查请求，不会创建实例。检查项包括是否填写了必需参数、请求格式、业务限制和ECS库存。如果检查不通过，则返回对应错误。如果检查通过，则返回DryRunOperation错误。
	// false：发送正常请求，通过检查后直接创建实例。
	dryRun := true
	if dryRun {
		fmt.Println("Performing dry run to validate parameters...")
	} else {
		fmt.Println("Creating instance...")
	}

	if err != nil {
		log.Fatalf("create client failed: %v", err)
	}
	// 查询可用镜像列表
	images, err := aliyun.DescribeImages(client)
	if err != nil {
		log.Fatalf("describe images failed: %v", err)
	}
	// 查询实例规格列表
	instanceTypes, err := aliyun.DescribeInstanceTypes(client)
	if err != nil {
		log.Fatalf("describe instance types failed: %v", err)
	}
	if len(instanceTypes) == 0 {
		log.Fatalf("no instance types available")
	}
	// 查询可用区列表
	zones, err := aliyun.DescribeAvailableResource(client, "InstanceType", "optimized", instanceTypes[0])
	if err != nil {
		log.Fatalf("describe zones failed: %v", err)
	}
	if len(zones) == 0 {
		log.Fatalf("no zones available for instance type %s", instanceTypes[0])
	}
	// 查询系统盘类型
	systemDiskCategories, err := aliyun.DescribeSystemDiskCategories(client, instanceTypes[0], zones[0])
	if err != nil {
		log.Fatalf("describe system disk categories failed: %v", err)
	}
	if len(systemDiskCategories) == 0 {
		log.Fatalf("no system disk category available for instance type %s in zone %s", instanceTypes[0], zones[0])
	}
	// 创建实例
	instanceId, err := aliyun.CreateInstance(client, images[0], "ecs.e-c1m2.xlarge", "", "cloud_essd")
	if err != nil {
		log.Fatalf("create instance failed: %v", err)
	}
	if err := aliyun.WaitForInstanceStatus(client, instanceId, "Stopped", 10*time.Minute, 5*time.Second); err != nil {
		log.Fatalf("wait for instance stopped failed: %v", err)
	}
	// 启动实例
	if err := aliyun.StartInstance(client, instanceId); err != nil {
		log.Fatalf("start instance failed: %v", err)
	}

	fmt.Printf("instance created and started: %s\n", instanceId)
}
