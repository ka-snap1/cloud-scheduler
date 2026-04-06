package compute

import (
	"context"
	"testing"
)

type serviceFakeAdapter struct {
	createCalled bool
	startCalled  bool
	stopCalled   bool
	deleteCalled bool
	statusCalled bool
}

func (a *serviceFakeAdapter) Name() string { return "aliyun" }

func (a *serviceFakeAdapter) ValidateProfile(profile *Profile) error {
	return profile.Validate()
}

func (a *serviceFakeAdapter) Create(_ context.Context, _ ComputeSpaceSpec, _ *Profile) (CreateResult, error) {
	a.createCalled = true
	return CreateResult{
		ProviderID: "i-123",
		Status: ComputeSpaceStatus{
			Phase:      string(PhaseProvisioning),
			Provider:   "aliyun",
			ProviderID: "i-123",
		},
	}, nil
}

func (a *serviceFakeAdapter) Start(_ context.Context, _ string) error {
	a.startCalled = true
	return nil
}

func (a *serviceFakeAdapter) Stop(_ context.Context, _ string) error {
	a.stopCalled = true
	return nil
}

func (a *serviceFakeAdapter) Delete(_ context.Context, _ string) error {
	a.deleteCalled = true
	return nil
}

func (a *serviceFakeAdapter) Status(_ context.Context, providerID string) (ComputeSpaceStatus, error) {
	a.statusCalled = true
	return ComputeSpaceStatus{Phase: string(PhaseRunning), ProviderID: providerID, Provider: "aliyun"}, nil
}

func TestNewService_RequiresDependencies(t *testing.T) {
	registry := NewAdapterRegistry()
	resolver := NewStaticProfileResolver([]Profile{})

	if _, err := NewService(nil, registry); err == nil {
		t.Fatalf("expected error when profile resolver is nil")
	}
	if _, err := NewService(resolver, nil); err == nil {
		t.Fatalf("expected error when adapter registry is nil")
	}
}

func TestService_Provision_RequiresProfileRef(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter := &serviceFakeAdapter{}
	if err := registry.Register(adapter); err != nil {
		t.Fatalf("register adapter failed: %v", err)
	}

	resolver := NewStaticProfileResolver([]Profile{{Name: "dev", Provider: "aliyun"}})
	svc, err := NewService(resolver, registry)
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}

	if _, err := svc.Provision(context.Background(), ComputeSpaceSpec{}); err == nil {
		t.Fatalf("expected error when profile_ref is missing")
	}
}

func TestService_Provision_Success(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter := &serviceFakeAdapter{}
	if err := registry.Register(adapter); err != nil {
		t.Fatalf("register adapter failed: %v", err)
	}

	resolver := NewStaticProfileResolver([]Profile{{
		Name:     "dev",
		Provider: "aliyun",
		Resources: ResourceProfile{
			ImageID:            "img-1",
			InstanceType:       "ecs.g6.large",
			SystemDiskCategory: "cloud_essd",
		},
	}})

	svc, err := NewService(resolver, registry)
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}

	result, err := svc.Provision(context.Background(), ComputeSpaceSpec{ProfileRef: "dev"})
	if err != nil {
		t.Fatalf("provision failed: %v", err)
	}
	if result.ProviderID != "i-123" {
		t.Fatalf("unexpected provider id: %s", result.ProviderID)
	}
	if !adapter.createCalled {
		t.Fatalf("expected create to be called")
	}
}

func TestService_DispatchLifecycleCalls(t *testing.T) {
	registry := NewAdapterRegistry()
	adapter := &serviceFakeAdapter{}
	if err := registry.Register(adapter); err != nil {
		t.Fatalf("register adapter failed: %v", err)
	}

	resolver := NewStaticProfileResolver([]Profile{{Name: "dev", Provider: "aliyun"}})
	svc, err := NewService(resolver, registry)
	if err != nil {
		t.Fatalf("new service failed: %v", err)
	}

	if err := svc.Start(context.Background(), "aliyun", "i-1"); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := svc.Stop(context.Background(), "aliyun", "i-1"); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	if err := svc.Delete(context.Background(), "aliyun", "i-1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	status, err := svc.Status(context.Background(), "aliyun", "i-1")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}

	if !adapter.startCalled || !adapter.stopCalled || !adapter.deleteCalled || !adapter.statusCalled {
		t.Fatalf("expected lifecycle methods to be dispatched")
	}
	if status.ProviderID != "i-1" {
		t.Fatalf("unexpected status provider id: %s", status.ProviderID)
	}
}
