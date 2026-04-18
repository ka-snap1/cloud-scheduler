package volcengine

import (
	"fmt"
	"strings"
	"time"

	"github.com/volcengine/volcengine-go-sdk/service/ecs"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

type CreateInstanceOptions struct {
	RegionID           string
	ZoneID             string
	ImageID            string
	InstanceTypeID     string
	SystemDiskCategory string
	SecurityGroupID    string
	SubnetID           string
	InstanceName       string
	Password           string
	KeyPairName        string
}

func mapSystemDiskCategory(category string) string {
	value := strings.ToUpper(strings.TrimSpace(category))
	switch value {
	case "", "CLOUD_ESSD", "ESSD", "ESSD_PL0":
		return "ESSD_PL0"
	case "ESSD_PL1", "ESSD_PL2", "ESSD_FLEXPL":
		return value
	case "CLOUD_SSD", "SSD":
		return "SSD"
	case "CLOUD_EFFICIENCY", "HDD":
		return "HDD"
	default:
		return value
	}
}

func CreateInstanceWithOptions(client *ecs.ECS, opts CreateInstanceOptions) (string, error) {
	if client == nil {
		return "", fmt.Errorf("ecs client is nil")
	}
	if strings.TrimSpace(opts.ZoneID) == "" {
		return "", fmt.Errorf("zone_id is required")
	}
	if strings.TrimSpace(opts.ImageID) == "" {
		return "", fmt.Errorf("image_id is required")
	}
	if strings.TrimSpace(opts.InstanceTypeID) == "" {
		return "", fmt.Errorf("instance_type is required")
	}
	if strings.TrimSpace(opts.SecurityGroupID) == "" {
		return "", fmt.Errorf("security_group_id is required")
	}
	if strings.TrimSpace(opts.SubnetID) == "" {
		return "", fmt.Errorf("subnet_id is required")
	}

	instanceName := strings.TrimSpace(opts.InstanceName)
	if instanceName == "" {
		instanceName = fmt.Sprintf("compute-go-%d", time.Now().Unix())
	}

	diskType := mapSystemDiskCategory(opts.SystemDiskCategory)
	request := &ecs.RunInstancesInput{
		Count:          volcengine.Int32(1),
		ImageId:        volcengine.String(strings.TrimSpace(opts.ImageID)),
		InstanceName:   volcengine.String(instanceName),
		InstanceTypeId: volcengine.String(strings.TrimSpace(opts.InstanceTypeID)),
		NetworkInterfaces: []*ecs.NetworkInterfaceForRunInstancesInput{
			{
				SecurityGroupIds: volcengine.StringSlice([]string{strings.TrimSpace(opts.SecurityGroupID)}),
				SubnetId:         volcengine.String(strings.TrimSpace(opts.SubnetID)),
			},
		},
		Volumes: []*ecs.VolumeForRunInstancesInput{
			{
				Size:       volcengine.Int32(40),
				VolumeType: volcengine.String(diskType),
			},
		},
		ZoneId: volcengine.String(strings.TrimSpace(opts.ZoneID)),
	}

	if password := strings.TrimSpace(opts.Password); password != "" {
		request.Password = volcengine.String(password)
	}
	if keyPairName := strings.TrimSpace(opts.KeyPairName); keyPairName != "" {
		request.KeyPairName = volcengine.String(keyPairName)
	}

	response, err := client.RunInstances(request)
	if err != nil {
		return "", err
	}
	if response == nil || len(response.InstanceIds) == 0 || response.InstanceIds[0] == nil {
		return "", fmt.Errorf("run instances succeeded but no instance id returned")
	}
	return volcengine.StringValue(response.InstanceIds[0]), nil
}

func StartInstance(client *ecs.ECS, instanceID string) error {
	if client == nil {
		return fmt.Errorf("ecs client is nil")
	}
	_, err := client.StartInstance(&ecs.StartInstanceInput{InstanceId: volcengine.String(strings.TrimSpace(instanceID))})
	return err
}

func StopInstance(client *ecs.ECS, instanceID string) error {
	if client == nil {
		return fmt.Errorf("ecs client is nil")
	}
	_, err := client.StopInstance(&ecs.StopInstanceInput{InstanceId: volcengine.String(strings.TrimSpace(instanceID))})
	return err
}

func DeleteInstance(client *ecs.ECS, instanceID string) error {
	if client == nil {
		return fmt.Errorf("ecs client is nil")
	}
	_, err := client.DeleteInstance(&ecs.DeleteInstanceInput{InstanceId: volcengine.String(strings.TrimSpace(instanceID))})
	return err
}

func DescribeInstance(client *ecs.ECS, instanceID string) (*ecs.InstanceForDescribeInstancesOutput, error) {
	if client == nil {
		return nil, fmt.Errorf("ecs client is nil")
	}

	response, err := client.DescribeInstances(&ecs.DescribeInstancesInput{
		InstanceIds: volcengine.StringSlice([]string{strings.TrimSpace(instanceID)}),
	})
	if err != nil {
		return nil, err
	}
	if response == nil || len(response.Instances) == 0 || response.Instances[0] == nil {
		return nil, fmt.Errorf("instance %s not found", instanceID)
	}
	return response.Instances[0], nil
}

func DescribeInstanceStatus(client *ecs.ECS, instanceID string) (string, string, string, error) {
	instance, err := DescribeInstance(client, instanceID)
	if err != nil {
		return "", "", "", err
	}

	status := strings.TrimSpace(volcengine.StringValue(instance.Status))
	publicIP := ""
	if instance.EipAddress != nil {
		publicIP = strings.TrimSpace(volcengine.StringValue(instance.EipAddress.IpAddress))
	}

	privateIP := ""
	if len(instance.NetworkInterfaces) > 0 && instance.NetworkInterfaces[0] != nil {
		privateIP = strings.TrimSpace(volcengine.StringValue(instance.NetworkInterfaces[0].PrimaryIpAddress))
	}

	return status, privateIP, publicIP, nil
}
