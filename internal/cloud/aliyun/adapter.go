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

type instancePreset struct {
	instanceType      string
	defaultDisk       string
	diskCategoryOrder []string
}

var describeSystemDiskCategoriesFn = DescribeSystemDiskCategories

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
	if strings.TrimSpace(profile.ZoneID) == "" {
		return fmt.Errorf("profile.zone_id is required")
	}
	if strings.TrimSpace(profile.Network.SecurityGroupID) == "" {
		return fmt.Errorf("profile.network.security_group_id is required")
	}
	if strings.TrimSpace(profile.Network.VSwitchID) == "" {
		return fmt.Errorf("profile.network.vswitch_id is required")
	}
	return nil
}

func (a *Adapter) Create(_ context.Context, spec compute.ComputeSpaceSpec, profile *compute.Profile) (compute.CreateResult, error) {
	if err := a.ValidateProfile(profile); err != nil {
		return compute.CreateResult{}, err
	}

	preset, ok := resolvePreset(spec)
	if !ok {
		return compute.CreateResult{}, fmt.Errorf("unable to map computespace spec to aliyun instance preset")
	}

	imageID := strings.TrimSpace(spec.OSImage)
	if imageID == "" {
		imageID = strings.TrimSpace(profile.Resources.ImageID)
	}
	if imageID == "" {
		return compute.CreateResult{}, fmt.Errorf("spec.os_image is required")
	}

	instanceType := preset.instanceType
	systemDiskCategory := chooseDiskCategory(
		preset.defaultDisk,
		preset.diskCategoryOrder,
		instanceType,
		profile.ZoneID,
		a.client,
	)

	instanceID, err := CreateInstanceWithOptions(
		a.client,
		CreateInstanceOptions{
			RegionID:           profile.RegionID,
			ZoneID:             profile.ZoneID,
			ImageID:            imageID,
			InstanceType:       instanceType,
			SystemDiskCategory: systemDiskCategory,
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

func resolvePreset(spec compute.ComputeSpaceSpec) (instancePreset, bool) {
	if !hasSizingHints(spec) {
		return instancePreset{}, false
	}

	if spec.GPU >= 2 {
		return instancePreset{
			instanceType:      "ecs.gn7i-c16g2.4xlarge",
			defaultDisk:       "cloud_essd",
			diskCategoryOrder: []string{"cloud_essd", "cloud_ssd", "cloud_efficiency"},
		}, true
	}

	if spec.GPU == 1 {
		model := strings.ToLower(strings.TrimSpace(spec.GPUModel))
		if model == "" || strings.Contains(model, "t4") || strings.Contains(model, "a10") || strings.Contains(model, "l20") {
			return instancePreset{
				instanceType:      "ecs.gn6i-c8g1.2xlarge",
				defaultDisk:       "cloud_essd",
				diskCategoryOrder: []string{"cloud_essd", "cloud_ssd", "cloud_efficiency"},
			}, true
		}
		return instancePreset{
			instanceType:      "ecs.gn6i-c8g1.2xlarge",
			defaultDisk:       "cloud_essd",
			diskCategoryOrder: []string{"cloud_essd", "cloud_ssd", "cloud_efficiency"},
		}, true
	}

	if spec.CPU <= 2 && spec.MemoryGB <= 4 {
		return instancePreset{
			instanceType:      "ecs.t6-c1m2.large",
			defaultDisk:       "cloud_efficiency",
			diskCategoryOrder: []string{"cloud_efficiency", "cloud_ssd", "cloud_essd"},
		}, true
	}
	if spec.CPU <= 4 && spec.MemoryGB <= 16 {
		return instancePreset{
			instanceType:      "ecs.g7.large",
			defaultDisk:       "cloud_essd",
			diskCategoryOrder: []string{"cloud_essd", "cloud_ssd", "cloud_efficiency"},
		}, true
	}
	if spec.CPU <= 8 && spec.MemoryGB <= 32 {
		return instancePreset{
			instanceType:      "ecs.c7.xlarge",
			defaultDisk:       "cloud_essd",
			diskCategoryOrder: []string{"cloud_essd", "cloud_ssd", "cloud_efficiency"},
		}, true
	}
	if spec.CPU <= 16 && spec.MemoryGB <= 64 {
		return instancePreset{
			instanceType:      "ecs.r7.2xlarge",
			defaultDisk:       "cloud_essd",
			diskCategoryOrder: []string{"cloud_essd", "cloud_ssd", "cloud_efficiency"},
		}, true
	}

	return instancePreset{}, false
}

func hasSizingHints(spec compute.ComputeSpaceSpec) bool {
	return spec.CPU > 0 || spec.GPU > 0 || spec.MemoryGB > 0 || strings.TrimSpace(spec.GPUModel) != ""
}

func chooseDiskCategory(current string, preferred []string, instanceType string, zoneID string, client *ecs.Client) string {
	if strings.TrimSpace(instanceType) == "" || strings.TrimSpace(zoneID) == "" || client == nil {
		return current
	}

	supported, err := describeSystemDiskCategoriesFn(client, instanceType, zoneID)
	if err != nil {
		return current
	}

	choice := chooseDiskCategoryFromSupported(current, preferred, supported)
	if choice == "" {
		return current
	}
	return choice
}

func chooseDiskCategoryFromSupported(current string, preferred []string, supported []string) string {
	supportedSet := make(map[string]struct{}, len(supported))
	for _, category := range supported {
		key := strings.ToLower(strings.TrimSpace(category))
		if key == "" {
			continue
		}
		supportedSet[key] = struct{}{}
	}

	for _, candidate := range buildCandidateOrder(current, preferred) {
		if _, ok := supportedSet[candidate]; ok {
			return candidate
		}
	}

	return ""
}

func buildCandidateOrder(current string, preferred []string) []string {
	seen := make(map[string]struct{})
	var result []string

	appendCandidate := func(value string) {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}

	appendCandidate(current)
	for _, candidate := range preferred {
		appendCandidate(candidate)
	}

	return result
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

// 根据computespace规格创建特定几种实例

var _ compute.Adapter = (*Adapter)(nil)
