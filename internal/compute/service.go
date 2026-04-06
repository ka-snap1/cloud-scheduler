package compute

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	profiles ProfileResolver
	adapters *AdapterRegistry
}

func NewService(profiles ProfileResolver, adapters *AdapterRegistry) (*Service, error) {
	if profiles == nil {
		return nil, fmt.Errorf("profile resolver is required")
	}
	if adapters == nil {
		return nil, fmt.Errorf("adapter registry is required")
	}
	return &Service{profiles: profiles, adapters: adapters}, nil
}

func (s *Service) Provision(ctx context.Context, spec ComputeSpaceSpec) (CreateResult, error) {
	profile, adapter, err := s.resolve(spec)
	if err != nil {
		return CreateResult{}, err
	}
	if err := adapter.ValidateProfile(profile); err != nil {
		return CreateResult{}, err
	}
	return adapter.Create(ctx, spec, profile)
}

func (s *Service) Start(ctx context.Context, provider string, providerID string) error {
	adapter, err := s.adapters.Resolve(provider)
	if err != nil {
		return err
	}
	return adapter.Start(ctx, providerID)
}

func (s *Service) Stop(ctx context.Context, provider string, providerID string) error {
	adapter, err := s.adapters.Resolve(provider)
	if err != nil {
		return err
	}
	return adapter.Stop(ctx, providerID)
}

func (s *Service) Delete(ctx context.Context, provider string, providerID string) error {
	adapter, err := s.adapters.Resolve(provider)
	if err != nil {
		return err
	}
	return adapter.Delete(ctx, providerID)
}

func (s *Service) Status(ctx context.Context, provider string, providerID string) (ComputeSpaceStatus, error) {
	adapter, err := s.adapters.Resolve(provider)
	if err != nil {
		return ComputeSpaceStatus{}, err
	}
	return adapter.Status(ctx, providerID)
}

func (s *Service) resolve(spec ComputeSpaceSpec) (*Profile, Adapter, error) {
	if strings.TrimSpace(spec.ProfileRef) == "" {
		return nil, nil, fmt.Errorf("spec.profile_ref is required")
	}

	profile, err := s.profiles.Resolve(context.Background(), spec.ProfileRef)
	if err != nil {
		return nil, nil, err
	}
	if err := profile.Validate(); err != nil {
		return nil, nil, err
	}

	provider := strings.TrimSpace(spec.Provider)
	if provider == "" {
		provider = profile.Provider
	}
	adapter, err := s.adapters.Resolve(provider)
	if err != nil {
		return nil, nil, err
	}
	return profile, adapter, nil
}
