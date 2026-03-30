package aliyun

import (
	"os"

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

func CreateInstance(client *ecs.Client) (string, error) {
	request := &ecs.CreateInstanceRequest{
		RegionId:        tea.String("cn-hangzhou"),
		ImageId:         tea.String("ubuntu_20_04_x64_20G_alibase_20220309.vhd"),
		InstanceType:    tea.String("ecs.t5-lc2m1.nano"),
		SecurityGroupId: tea.String("sg-uf6gqj8n7l0b4h9z"),
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
