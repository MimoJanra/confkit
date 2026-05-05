package aws

import (
	"testing"

	"confkit"
)

func TestNewRegionFailoverSource(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2"}
	src, err := NewRegionFailoverSource(regions, func(region string) confkit.Source {
		return confkit.NewErrorSource(nil)
	})

	if err != nil {
		t.Fatalf("NewRegionFailoverSource failed: %v", err)
	}

	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if len(src.regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(src.regions))
	}
}

func TestRegionFailoverSourceName(t *testing.T) {
	src := &RegionFailoverSource{}
	name := src.Name()
	if name != "multiregion" {
		t.Errorf("Expected name 'multiregion', got %q", name)
	}
}

func TestRegionFailoverSourceEmptyRegions(t *testing.T) {
	_, err := NewRegionFailoverSource([]string{}, func(region string) confkit.Source {
		return confkit.NewErrorSource(nil)
	})

	if err == nil {
		t.Fatal("Expected error for empty regions")
	}
}

func TestRegionFailoverSourceGetCurrentRegion(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	src, _ := NewRegionFailoverSource(regions, func(region string) confkit.Source {
		return confkit.NewErrorSource(nil)
	})

	current := src.GetCurrentRegion()
	if current != "us-east-1" {
		t.Errorf("Expected us-east-1, got %s", current)
	}
}

func TestRegionFailoverSourceGetHealthyRegions(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2"}
	src, _ := NewRegionFailoverSource(regions, func(region string) confkit.Source {
		return confkit.NewErrorSource(nil)
	})

	healthy := src.GetHealthyRegions()
	if len(healthy) != 0 {
		t.Errorf("Expected no healthy regions initially, got %d", len(healthy))
	}
}

func TestFromAWSSecretsManagerMultiRegion(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2"}
	src := FromAWSSecretsManagerMultiRegion("myapp/db", regions)

	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "multiregion" && src.Name() != "error" {
		t.Errorf("Expected multiregion or error, got %q", src.Name())
	}
}

func TestFromAWSSSMParameterStoreMultiRegion(t *testing.T) {
	regions := []string{"us-east-1", "us-west-2"}
	src := FromAWSSSMParameterStoreMultiRegion("/myapp/", regions)

	if src == nil {
		t.Fatal("Expected non-nil source")
	}

	if src.Name() != "multiregion" && src.Name() != "error" {
		t.Errorf("Expected multiregion or error, got %q", src.Name())
	}
}

func TestRegionFailoverSourceCurrentRegionUpdate(t *testing.T) {
	regions := []string{"region-a", "region-b"}
	src, _ := NewRegionFailoverSource(regions, func(region string) confkit.Source {
		return confkit.NewErrorSource(nil)
	})

	if src.GetCurrentRegion() != "region-a" {
		t.Error("Expected initial region to be region-a")
	}

	src.currentRegion = 1
	if src.GetCurrentRegion() != "region-b" {
		t.Error("Expected current region to be region-b after update")
	}
}

func TestRegionFailoverSourceRegionCache(t *testing.T) {
	src := &RegionFailoverSource{
		regionCache: make(map[string]bool),
	}

	src.updateRegionCache("us-east-1", true)
	src.updateRegionCache("us-west-2", false)

	if len(src.regionCache) != 2 {
		t.Errorf("Expected 2 entries in cache, got %d", len(src.regionCache))
	}

	if !src.regionCache["us-east-1"] {
		t.Error("Expected us-east-1 to be healthy")
	}

	if src.regionCache["us-west-2"] {
		t.Error("Expected us-west-2 to be unhealthy")
	}
}
