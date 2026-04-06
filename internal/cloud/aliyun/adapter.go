package aliyun

import (
	"context"
	"fmt"
	"strings"

	"cloud-scheduler/internal/compute"

	ecs "github.com/alibabacloud-go/ecs-20140526/v4/client"
)

const ProviderName = "aliyun"

type Adapter struct {
	client *ecs.Client
}

func NewAdapter(client *ecs.Client) (*Adapter, error) {
	if client == nil {
		return nil, fmt.Errorf("aliyun client is nil")
	}
	return &Adapter{client: client}, nil
}

func NewAdapterFromEnv() (*Adapter, error) {
	client, err := CreateClient()
	if err != nil {
		return nil, err
	}
	return NewAdapter(client)
}

func (a *Adapter) Name() string {
	return ProviderName
}

func (a *Adapter) ValidateProfile(profile *compute.Profile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}
	if !strings.EqualFold(strings.TrimSpace(profile.Provider), ProviderName) {
		return fmt.Errorf("profile provider must be %s", ProviderName)
	}
	if strings.TrimSpace(profile.Resources.ImageID) == "" {
		return fmt.Errorf("profile.resources.image_id is required")
	}
	if strings.TrimSpace(profile.Resources.InstanceType) == "" {
		return fmt.Errorf("profile.resources.instance_type is required")
	}
	if strings.TrimSpace(profile.Resources.SystemDiskCategory) == "" {
		return fmt.Errorf("profile.resources.system_disk_category is required")
	}
	return nil
}

func (a *Adapter) Create(_ context.Context, _ compute.ComputeSpaceSpec, profile *compute.Profile) (compute.CreateResult, error) {
	if err := a.ValidateProfile(profile); err != nil {
		return compute.CreateResult{}, err
	}

	instanceID, err := CreateInstanceWithOptions(
		a.client,
		CreateInstanceOptions{
			RegionID:           profile.RegionID,
			ZoneID:             profile.ZoneID,
			ImageID:            profile.Resources.ImageID,
			InstanceType:       profile.Resources.InstanceType,
			SystemDiskCategory: profile.Resources.SystemDiskCategory,
			SecurityGroupID:    profile.Network.SecurityGroupID,
			VSwitchID:          profile.Network.VSwitchID,
		},
	)
	if err != nil {
		return compute.CreateResult{}, err
	}

	return compute.CreateResult{
		ProviderID: instanceID,
		Status: compute.ComputeSpaceStatus{
			Phase:      string(compute.PhaseProvisioning),
			Provider:   ProviderName,
			ProviderID: instanceID,
			SSHuser:    profile.SSH.User,
		},
	}, nil
}

func (a *Adapter) Start(_ context.Context, providerID string) error {
	return StartInstance(a.client, providerID)
}

func (a *Adapter) Stop(_ context.Context, providerID string) error {
	return StopInstance(a.client, providerID)
}

func (a *Adapter) Delete(_ context.Context, providerID string) error {
	return DeleteInstance(a.client, providerID)
}

func (a *Adapter) Status(_ context.Context, providerID string) (compute.ComputeSpaceStatus, error) {
	status, err := DescribeInstanceStatus(a.client, providerID)
	if err != nil {
		return compute.ComputeSpaceStatus{}, err
	}

	phase := compute.PhasePending
	switch status {
	case "Running":
		phase = compute.PhaseRunning
	case "Stopped":
		phase = compute.PhaseStopped
	case "Stopping":
		phase = compute.PhaseStopping
	case "Starting":
		phase = compute.PhaseStarting
	}

	return compute.ComputeSpaceStatus{
		Phase:      string(phase),
		Provider:   ProviderName,
		ProviderID: providerID,
	}, nil
}

var _ compute.Adapter = (*Adapter)(nil)
