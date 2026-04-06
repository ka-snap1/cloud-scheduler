package compute

import (
	"context"
	"testing"
)

type registryFakeAdapter struct{}

func (a *registryFakeAdapter) Name() string { return "AliYun" }

func (a *registryFakeAdapter) ValidateProfile(_ *Profile) error { return nil }

func (a *registryFakeAdapter) Create(_ context.Context, _ ComputeSpaceSpec, _ *Profile) (CreateResult, error) {
	return CreateResult{}, nil
}

func (a *registryFakeAdapter) Start(_ context.Context, _ string) error { return nil }

func (a *registryFakeAdapter) Stop(_ context.Context, _ string) error { return nil }

func (a *registryFakeAdapter) Delete(_ context.Context, _ string) error { return nil }

func (a *registryFakeAdapter) Status(_ context.Context, _ string) (ComputeSpaceStatus, error) {
	return ComputeSpaceStatus{}, nil
}

func TestAdapterRegistry_RegisterAndResolve(t *testing.T) {
	r := NewAdapterRegistry()

	if err := r.Register(nil); err == nil {
		t.Fatalf("expected error when registering nil adapter")
	}

	adapter := &registryFakeAdapter{}
	if err := r.Register(adapter); err != nil {
		t.Fatalf("register adapter failed: %v", err)
	}

	resolved, err := r.Resolve("aliyun")
	if err != nil {
		t.Fatalf("resolve adapter failed: %v", err)
	}
	if resolved != adapter {
		t.Fatalf("resolved adapter mismatch")
	}

	if _, err := r.Resolve(""); err == nil {
		t.Fatalf("expected error when resolving empty provider")
	}

	if _, err := r.Resolve("unknown"); err == nil {
		t.Fatalf("expected error when resolving unknown provider")
	}
}
