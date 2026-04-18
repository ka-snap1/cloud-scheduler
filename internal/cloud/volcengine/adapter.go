package volcengine

import (
	"context"
	"fmt"
	"strings"

	"cloud-scheduler/internal/compute"

	"github.com/volcengine/volcengine-go-sdk/service/ecs"
)

const ProviderName = "volcengine"

type Adapter struct {
	client *ecs.ECS
}

func NewAdapter(client *ecs.ECS) (*Adapter, error) {
	if client == nil {
		return nil, fmt.Errorf("volcengine client is nil")
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
	if strings.TrimSpace(profile.ZoneID) == "" {
		return fmt.Errorf("profile.zone_id is required")
	}
	if strings.TrimSpace(profile.Network.SecurityGroupID) == "" {
		return fmt.Errorf("profile.network.security_group_id is required")
	}
	if strings.TrimSpace(profile.Network.VSwitchID) == "" {
		return fmt.Errorf("profile.network.vswitch_id is required (used as subnet_id for volcengine)")
	}
	return nil
}

type instancePreset struct {
	instanceTypeID string
	diskType       string
}

func resolvePreset(spec compute.ComputeSpaceSpec) (instancePreset, bool) {
	if spec.GPU >= 2 {
		return instancePreset{instanceTypeID: "ecs.g1ie.2xlarge", diskType: "ESSD_PL0"}, true
	}
	if spec.GPU == 1 {
		return instancePreset{instanceTypeID: "ecs.g1ie.xlarge", diskType: "ESSD_PL0"}, true
	}
	if spec.CPU <= 2 && spec.MemoryGB <= 4 {
		return instancePreset{instanceTypeID: "ecs.t1.small", diskType: "ESSD_PL0"}, true
	}
	if spec.CPU <= 4 && spec.MemoryGB <= 8 {
		return instancePreset{instanceTypeID: "ecs.g1.small", diskType: "ESSD_PL0"}, true
	}
	if spec.CPU <= 8 && spec.MemoryGB <= 16 {
		return instancePreset{instanceTypeID: "ecs.g1.medium", diskType: "ESSD_PL0"}, true
	}
	return instancePreset{instanceTypeID: "ecs.g1.large", diskType: "ESSD_PL0"}, true
}

func (a *Adapter) Create(_ context.Context, spec compute.ComputeSpaceSpec, profile *compute.Profile) (compute.CreateResult, error) {
	if err := a.ValidateProfile(profile); err != nil {
		return compute.CreateResult{}, err
	}

	preset, _ := resolvePreset(spec)
	instanceType := strings.TrimSpace(profile.Resources.InstanceType)
	if instanceType == "" {
		instanceType = preset.instanceTypeID
	}

	imageID := strings.TrimSpace(spec.OSImage)
	if imageID == "" {
		imageID = strings.TrimSpace(profile.Resources.ImageID)
	}
	if imageID == "" {
		return compute.CreateResult{}, fmt.Errorf("spec.os_image is required")
	}

	diskCategory := strings.TrimSpace(profile.Resources.SystemDiskCategory)
	if diskCategory == "" {
		diskCategory = preset.diskType
	}

	instanceID, err := CreateInstanceWithOptions(
		a.client,
		CreateInstanceOptions{
			RegionID:           profile.RegionID,
			ZoneID:             profile.ZoneID,
			ImageID:            imageID,
			InstanceTypeID:     instanceType,
			SystemDiskCategory: diskCategory,
			SecurityGroupID:    profile.Network.SecurityGroupID,
			SubnetID:           profile.Network.VSwitchID,
			InstanceName:       fmt.Sprintf("compute-go-%s", strings.ReplaceAll(strings.ToLower(strings.TrimSpace(profile.Name)), " ", "-")),
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

func mapPhase(status string) compute.Phase {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running":
		return compute.PhaseRunning
	case "stopped":
		return compute.PhaseStopped
	case "stopping":
		return compute.PhaseStopping
	case "starting":
		return compute.PhaseStarting
	case "deleting":
		return compute.PhaseDeleting
	case "deleted":
		return compute.PhaseDeleted
	case "error", "failed":
		return compute.PhaseFailed
	default:
		return compute.PhasePending
	}
}

func (a *Adapter) Status(_ context.Context, providerID string) (compute.ComputeSpaceStatus, error) {
	status, privateIP, publicIP, err := DescribeInstanceStatus(a.client, providerID)
	if err != nil {
		return compute.ComputeSpaceStatus{}, err
	}

	return compute.ComputeSpaceStatus{
		Phase:      string(mapPhase(status)),
		Provider:   ProviderName,
		ProviderID: providerID,
		PrivateIP:  privateIP,
		PublicIP:   publicIP,
	}, nil
}

var _ compute.Adapter = (*Adapter)(nil)
