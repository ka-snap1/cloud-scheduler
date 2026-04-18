//go:build volcen

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud-scheduler/internal/cloud/volcengine"
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

func testVolcengineCreate() {
	ctx := context.Background()

	profilePath := getenvAny("VOLC_PROFILE_PATH")
	if profilePath == "" {
		profilePath = "internal/config/volcen_profiles.yaml"
	}

	profileName := getenvAny("VOLC_PROFILE_NAME")
	if profileName == "" {
		profileName = "test-volcen-gpu"
	}

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

	if !strings.EqualFold(strings.TrimSpace(profile.Provider), volcengine.ProviderName) {
		fmt.Printf("invalid profile provider=%q, expected %q\n", profile.Provider, volcengine.ProviderName)
		os.Exit(1)
	}

	adapter, err := volcengine.NewAdapterFromEnv()
	if err != nil {
		fmt.Printf("failed to create volcengine adapter: %v\n", err)
		os.Exit(1)
	}

	if err := adapter.ValidateProfile(profile); err != nil {
		fmt.Printf("profile validation failed: %v\n", err)
		os.Exit(1)
	}

	imageID := strings.TrimSpace(getenvAny("VOLC_IMAGE_ID"))
	if imageID == "" {
		imageID = strings.TrimSpace(profile.Resources.ImageID)
	}
	if imageID == "" {
		fmt.Println("image id is required: set VOLC_IMAGE_ID or profile.resources.image_id")
		os.Exit(1)
	}

	spec := compute.ComputeSpaceSpec{
		CPU:      8,
		GPU:      1,
		GPUModel: "T4",
		MemoryGB: 32,
		OSImage:  imageID,
	}

	result, err := adapter.Create(ctx, spec, profile)
	if err != nil {
		fmt.Printf("create instance failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("create succeeded, instance=%s, phase=%s\n", result.ProviderID, result.Status.Phase)

	status, err := adapter.Status(ctx, result.ProviderID)
	if err != nil {
		fmt.Printf("status query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("status=%s private_ip=%s public_ip=%s\n", status.Phase, status.PrivateIP, status.PublicIP)
}

func main() {
	testVolcengineCreate()
}
