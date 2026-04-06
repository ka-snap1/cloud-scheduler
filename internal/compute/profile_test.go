package compute

import (
	"context"
	"testing"
)

func TestProfileValidate(t *testing.T) {
	var nilProfile *Profile
	if err := nilProfile.Validate(); err == nil {
		t.Fatalf("expected error for nil profile")
	}

	p := &Profile{}
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error for empty profile")
	}

	p.Name = "dev"
	if err := p.Validate(); err == nil {
		t.Fatalf("expected error for missing provider")
	}

	p.Provider = "aliyun"
	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}

func TestStaticProfileResolver_Resolve(t *testing.T) {
	resolver := NewStaticProfileResolver([]Profile{
		{
			Name:     "Dev",
			Provider: "aliyun",
		},
	})

	profile, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("resolve profile failed: %v", err)
	}
	if profile == nil {
		t.Fatalf("resolved profile is nil")
	}
	if profile.Name != "Dev" {
		t.Fatalf("unexpected profile name: %s", profile.Name)
	}

	if _, err := resolver.Resolve(context.Background(), ""); err == nil {
		t.Fatalf("expected error for empty profile name")
	}

	if _, err := resolver.Resolve(context.Background(), "missing"); err == nil {
		t.Fatalf("expected error for missing profile")
	}
}
