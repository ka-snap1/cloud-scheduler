package compute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type profileListYAML struct {
	Profiles []Profile `yaml:"profiles"`
}

func LoadProfilesFromYAML(path string) ([]Profile, error) {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return nil, fmt.Errorf("yaml path is required")
	}

	content, err := os.ReadFile(filepath.Clean(cleanPath))
	if err != nil {
		return nil, fmt.Errorf("read yaml failed: %w", err)
	}

	var wrapped profileListYAML
	if err := yaml.Unmarshal(content, &wrapped); err != nil {
		return nil, fmt.Errorf("parse yaml failed: %w", err)
	}

	profiles := wrapped.Profiles
	if len(profiles) == 0 {
		if err := yaml.Unmarshal(content, &profiles); err != nil {
			return nil, fmt.Errorf("parse yaml failed: %w", err)
		}
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles found in yaml")
	}

	for i := range profiles {
		if err := profiles[i].Validate(); err != nil {
			return nil, fmt.Errorf("invalid profile at index %d: %w", i, err)
		}
	}

	return profiles, nil
}
