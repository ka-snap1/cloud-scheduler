package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud-scheduler/internal/cloud/aliyun"
	"cloud-scheduler/internal/compute"
)

func getenvAny(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func testAliyun() {
	ctx := context.Background()
	profilePath := "internal/config/aliyun_profiles.yaml"
	profileName := "test-aliyun-gpu"

	profiles, err := compute.LoadProfilesFromYAML(profilePath)
	if err != nil {
		fmt.Printf("load profiles from yaml failed: %v\n", err)
		os.Exit(1)
	}

	resolver := compute.NewStaticProfileResolver(profiles)
	profile, err := resolver.Resolve(ctx, profileName)
	if err != nil {
		fmt.Printf("resolve profile failed: %v\n", err)
		os.Exit(1)
	}

	adapter, err := aliyun.NewAdapterFromEnv()
	if err != nil {
		fmt.Printf("failed to create aliyun adapter: %v\n", err)
		os.Exit(1)
	}

	if err := adapter.ValidateProfile(profile); err != nil {
		fmt.Printf("profile validation failed: %v\n", err)
		os.Exit(1)
	}

	imageID := "ubuntu_18_04_64_20G_alibase_20190624.vhd"

	spec := compute.ComputeSpaceSpec{
		CPU:      8,
		GPU:      1,
		GPUModel: "T4",
		MemoryGB: 32,
		OSImage:  imageID,
	}

	result, err := adapter.Create(ctx, spec, profile)
	if err != nil {
		fmt.Printf("create gpu=1 instance failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("gpu=1 create succeeded, instance=%s, phase=%s\n", result.ProviderID, result.Status.Phase)
}

func main() {
	testAliyun()
}
