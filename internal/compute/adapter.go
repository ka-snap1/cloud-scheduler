package compute

import (
	"context"
	"fmt"
	"strings"
)

type Phase string

const (
	PhasePending      Phase = "Pending"
	PhaseProvisioning Phase = "Provisioning"
	PhaseStarting     Phase = "Starting"
	PhaseRunning      Phase = "Running"
	PhaseStopping     Phase = "Stopping"
	PhaseStopped      Phase = "Stopped"
	PhaseDeleting     Phase = "Deleting"
	PhaseDeleted      Phase = "Deleted"
	PhaseFailed       Phase = "Failed"
)

type Adapter interface {
	Name() string
	ValidateProfile(profile *Profile) error
	Create(ctx context.Context, spec ComputeSpaceSpec, profile *Profile) (CreateResult, error)
	Start(ctx context.Context, providerID string) error
	Stop(ctx context.Context, providerID string) error
	Delete(ctx context.Context, providerID string) error
	Status(ctx context.Context, providerID string) (ComputeSpaceStatus, error)
}

type CreateResult struct {
	ProviderID string
	Status     ComputeSpaceStatus
}

type AdapterRegistry struct {
	adapters map[string]Adapter
}

func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{adapters: make(map[string]Adapter)}
}

func (r *AdapterRegistry) Register(adapter Adapter) error {
	if adapter == nil {
		return fmt.Errorf("adapter is nil")
	}
	name := strings.TrimSpace(adapter.Name())
	if name == "" {
		return fmt.Errorf("adapter name is required")
	}
	r.adapters[strings.ToLower(name)] = adapter
	return nil
}

func (r *AdapterRegistry) Resolve(provider string) (Adapter, error) {
	key := strings.ToLower(strings.TrimSpace(provider))
	if key == "" {
		return nil, fmt.Errorf("provider is required")
	}
	adapter, ok := r.adapters[key]
	if !ok {
		return nil, fmt.Errorf("adapter not found for provider %s", provider)
	}
	return adapter, nil
}
