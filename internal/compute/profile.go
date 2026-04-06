package compute

import (
	"context"
	"fmt"
	"strings"
)

type Profile struct {
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	RegionID  string            `json:"region_id,omitempty"`
	ZoneID    string            `json:"zone_id,omitempty"`
	Resources ResourceProfile   `json:"resources"`
	Network   NetworkProfile    `json:"network"`
	SSH       SSHProfile        `json:"ssh"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type ResourceProfile struct {
	InstanceType       string `json:"instance_type,omitempty"`
	ImageID            string `json:"image_id,omitempty"`
	SystemDiskCategory string `json:"system_disk_category,omitempty"`
}

type NetworkProfile struct {
	SecurityGroupID string `json:"security_group_id,omitempty"`
	VSwitchID       string `json:"vswitch_id,omitempty"`
}

type SSHProfile struct {
	User string `json:"user,omitempty"`
	Port int    `json:"port,omitempty"`
}

func (p *Profile) Validate() error {
	if p == nil {
		return fmt.Errorf("profile is nil")
	}
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("profile name is required")
	}
	if strings.TrimSpace(p.Provider) == "" {
		return fmt.Errorf("profile provider is required")
	}
	return nil
}

type ProfileResolver interface {
	Resolve(ctx context.Context, name string) (*Profile, error)
}

type StaticProfileResolver struct {
	profiles map[string]Profile
}

func NewStaticProfileResolver(profiles []Profile) *StaticProfileResolver {
	stored := make(map[string]Profile, len(profiles))
	for _, profile := range profiles {
		stored[strings.ToLower(strings.TrimSpace(profile.Name))] = profile
	}
	return &StaticProfileResolver{profiles: stored}
}

func (r *StaticProfileResolver) Resolve(_ context.Context, name string) (*Profile, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return nil, fmt.Errorf("profile name is required")
	}
	profile, ok := r.profiles[key]
	if !ok {
		return nil, fmt.Errorf("profile %s not found", name)
	}
	copy := profile
	return &copy, nil
}
