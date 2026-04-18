package volcengine

import (
	"fmt"
	"os"
	"strings"

	"github.com/volcengine/volcengine-go-sdk/service/ecs"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
)

func getenvAny(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func CreateClient() (*ecs.ECS, error) {
	ak := getenvAny("VOLCENGINE_ACCESS_KEY")
	sk := getenvAny("VOLCENGINE_SECRET_KEY")
	region := getenvAny("VOLCENGINE_REGION", "VOLCENGINE_REGION_ID")

	if ak == "" {
		return nil, fmt.Errorf("VOLCENGINE_ACCESS_KEY (or VOLCENGINE_ACCESS_KEY_ID) is required")
	}
	if sk == "" {
		return nil, fmt.Errorf("VOLCENGINE_SECRET_KEY (or VOLCENGINE_ACCESS_KEY_SECRET) is required")
	}
	if region == "" {
		region = "cn-beijing"
	}

	return newECSClient(ak, sk, region)
}

func newECSClient(accessKeyID, accessKeySecret, region string) (*ecs.ECS, error) {
	creds := credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	config := volcengine.NewConfig().WithRegion(region).WithCredentials(creds)

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return ecs.New(sess), nil
}
