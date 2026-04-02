package aliyun

import (
	"fmt"
	"os"
	"time"

	ecs "github.com/alibabacloud-go/ecs-20140526/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

func getenvAny(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// 查询可用镜像列表
func DescribeImages(client *ecs.Client) ([]string, error) {
	request := &ecs.DescribeImagesRequest{
		RegionId:        tea.String("cn-hangzhou"),
		ImageOwnerAlias: tea.String("system"), // 只要官方镜像
	}
	response, err := client.DescribeImages(request)
	if err != nil {
		return nil, err
	}
	imageIds := make([]string, len(response.Body.Images.Image))
	for i, img := range response.Body.Images.Image {
		imageIds[i] = tea.StringValue(img.ImageId)
	}
	return imageIds, nil
}

// 查询实例规格列表
func DescribeInstanceTypes(client *ecs.Client) ([]string, error) {
	request := &ecs.DescribeInstanceTypesRequest{}
	response, err := client.DescribeInstanceTypes(request)
	if err != nil {
		return nil, err
	}
	instanceTypes := make([]string, len(response.Body.InstanceTypes.InstanceType))
	for i, it := range response.Body.InstanceTypes.InstanceType {
		instanceTypes[i] = tea.StringValue(it.InstanceTypeId)
	}
	return instanceTypes, nil
}

// 查询可用区列表 regionId: cn-hangzhou, instanceType: 实例规格
func DescribeAvailableResource(client *ecs.Client, destinationResource string, ioOptimized string, instanceType string) ([]string, error) {
	request := &ecs.DescribeAvailableResourceRequest{
		RegionId:            tea.String("cn-hangzhou"),
		DestinationResource: tea.String(destinationResource),
		IoOptimized:         tea.String(ioOptimized),
		InstanceType:        tea.String(instanceType),
	}
	response, err := client.DescribeAvailableResource(request)
	if err != nil {
		return nil, err
	}
	zones := make([]string, len(response.Body.AvailableZones.AvailableZone))
	for i, az := range response.Body.AvailableZones.AvailableZone {
		zones[i] = tea.StringValue(az.ZoneId)
	}
	return zones, nil
}

// 查询指定可用区支持的系统盘类型
func DescribeSystemDiskCategories(client *ecs.Client, instanceType string, zoneId string) ([]string, error) {
	request := &ecs.DescribeAvailableResourceRequest{
		RegionId:            tea.String("cn-hangzhou"),
		DestinationResource: tea.String("SystemDisk"),
		ZoneId:              tea.String(zoneId),
		InstanceType:        tea.String(instanceType),
	}
	response, err := client.DescribeAvailableResource(request)
	if err != nil {
		return nil, err
	}
	var categories []string
	for _, az := range response.Body.AvailableZones.AvailableZone {
		if tea.StringValue(az.ZoneId) != zoneId {
			continue
		}
		for _, res := range az.AvailableResources.AvailableResource {
			if tea.StringValue(res.Type) != "SystemDisk" {
				continue
			}
			for _, sr := range res.SupportedResources.SupportedResource {
				if tea.StringValue(sr.Status) == "Available" {
					categories = append(categories, tea.StringValue(sr.Value))
				}
			}
		}
	}
	if len(categories) == 0 {
		for _, az := range response.Body.AvailableZones.AvailableZone {
			if tea.StringValue(az.ZoneId) != zoneId {
				continue
			}
			for _, res := range az.AvailableResources.AvailableResource {
				if tea.StringValue(res.Type) != "SystemDisk" {
					continue
				}
				for _, sr := range res.SupportedResources.SupportedResource {
					categories = append(categories, tea.StringValue(sr.Value))
				}
			}
		}
	}
	return categories, nil
}

// 根据镜像ID创建实例
func CreateInstance(client *ecs.Client, imageId string, instanceType string, zoneId string, systemDiskCategory string) (string, error) {
	request := &ecs.CreateInstanceRequest{
		RegionId:     tea.String("cn-hangzhou"),
		ZoneId:       tea.String(zoneId),
		ImageId:      tea.String(imageId),
		InstanceType: tea.String(instanceType),
		SystemDisk: &ecs.CreateInstanceRequestSystemDisk{
			Category: tea.String(systemDiskCategory),
		},
		SecurityGroupId: tea.String("sg-bp104szh7za9nkpwpuyl"),
		VSwitchId:       tea.String("vsw-bp1eryqa4uwa037jht3le"),
	}
	response, err := client.CreateInstance(request)
	if err != nil {
		return "", err
	}
	return tea.StringValue(response.Body.InstanceId), nil
}

func StartInstance(client *ecs.Client, instanceId string) error {
	request := &ecs.StartInstanceRequest{
		InstanceId: tea.String(instanceId),
	}
	_, err := client.StartInstance(request)
	return err
}

func StopInstance(client *ecs.Client, instanceId string) error {
	request := &ecs.StopInstanceRequest{
		InstanceId: tea.String(instanceId),
	}
	_, err := client.StopInstance(request)
	return err
}

func DeleteInstance(client *ecs.Client, instanceId string) error {
	request := &ecs.DeleteInstanceRequest{
		InstanceId: tea.String(instanceId),
	}
	_, err := client.DeleteInstance(request)
	return err
}

func DescribeInstanceStatus(client *ecs.Client, instanceId string) (string, error) {
	request := &ecs.DescribeInstancesRequest{
		InstanceIds: tea.String(fmt.Sprintf("[\"%s\"]", instanceId)),
	}
	response, err := client.DescribeInstances(request)
	if err != nil {
		return "", err
	}
	if response.Body == nil || response.Body.Instances == nil || len(response.Body.Instances.Instance) == 0 {
		return "", fmt.Errorf("instance %s not found", instanceId)
	}
	return tea.StringValue(response.Body.Instances.Instance[0].Status), nil
}

func WaitForInstanceStatus(client *ecs.Client, instanceId string, wantStatus string, timeout time.Duration, interval time.Duration) error {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	for {
		status, err := DescribeInstanceStatus(client, instanceId)
		if err != nil {
			return err
		}
		if status == wantStatus {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for instance %s status %s, last status %s", instanceId, wantStatus, status)
		}
		time.Sleep(interval)
	}
}
