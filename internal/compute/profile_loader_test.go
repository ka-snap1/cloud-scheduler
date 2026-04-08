package compute

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfilesFromYAML_WrappedFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiles.yaml")

	content := `profiles:
  - name: test-aliyun
    provider: aliyun
    region_id: cn-hangzhou
    zone_id: cn-hangzhou-i
    resources:
      image_id: ubuntu_20_04_x64_20G_alibase_20230626.vhd
      instance_type: ecs.t5-lc2m1.nano
      system_disk_category: cloud_efficiency
    network:
      security_group_id: sg-xxx
      vswitch_id: vsw-xxx
    ssh:
      user: root
      port: 22
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write yaml failed: %v", err)
	}

	profiles, err := LoadProfilesFromYAML(path)
	if err != nil {
		t.Fatalf("LoadProfilesFromYAML failed: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("profiles length=%d, want 1", len(profiles))
	}
	if profiles[0].ZoneID != "cn-hangzhou-i" {
		t.Fatalf("zone id=%s, want cn-hangzhou-i", profiles[0].ZoneID)
	}
	if profiles[0].Resources.ImageID == "" {
		t.Fatalf("image id should not be empty")
	}
}

func TestLoadProfilesFromYAML_RequiresPath(t *testing.T) {
	if _, err := LoadProfilesFromYAML(" "); err == nil {
		t.Fatalf("expected error for empty path")
	}
}

func TestLoadProfilesFromYAML_RequiresProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte("profiles: []\n"), 0o644); err != nil {
		t.Fatalf("write yaml failed: %v", err)
	}
	if _, err := LoadProfilesFromYAML(path); err == nil {
		t.Fatalf("expected error for empty profiles")
	}
}
