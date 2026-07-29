package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/MimoJanra/confkit"
)

// RegionFailoverSource wraps one Source per AWS region and fails over between them.
//
// Lookups start at the region that last succeeded and try each remaining region in turn, so a
// regional outage costs one failed call rather than failing the load. Region health is
// recorded for inspection through GetHealthyRegions.
type RegionFailoverSource struct {
	regions       []string
	sources       []confkit.Source
	currentRegion int
	cacheMutex    sync.RWMutex
	regionCache   map[string]bool
	cacheTTL      time.Duration
	lastCacheAt   time.Time
}

// NewRegionFailoverSource builds one Source per region using sourceFactory, which is called
// once per region at construction. At least one region is required.
func NewRegionFailoverSource(regions []string, sourceFactory func(region string) confkit.Source) (*RegionFailoverSource, error) {
	if len(regions) == 0 {
		return nil, fmt.Errorf("must provide at least one region")
	}

	sources := make([]confkit.Source, len(regions))
	for i, region := range regions {
		sources[i] = sourceFactory(region)
	}

	return &RegionFailoverSource{
		regions:       regions,
		sources:       sources,
		currentRegion: 0,
		regionCache:   make(map[string]bool),
		cacheTTL:      5 * time.Minute,
	}, nil
}

// Name returns "multiregion".
func (r *RegionFailoverSource) Name() string {
	return "multiregion"
}

// Lookup tries each region in turn, beginning with the one that last succeeded, and remembers
// the region that answers.
//
// A region reporting the field as absent is treated as authoritative for that region and does
// not mark it unhealthy. If no region has the field, the last error is returned, or not-found
// if every region simply lacked it.
func (r *RegionFailoverSource) Lookup(ctx context.Context, field *confkit.FieldInfo) (any, bool, error) {
	r.cacheMutex.RLock()
	startRegion := r.currentRegion
	r.cacheMutex.RUnlock()

	var lastErr error
	for i := 0; i < len(r.regions); i++ {
		regionIdx := (startRegion + i) % len(r.regions)
		source := r.sources[regionIdx]

		value, ok, err := source.Lookup(ctx, field)
		if err == nil && ok {
			r.cacheMutex.Lock()
			r.currentRegion = regionIdx
			r.cacheMutex.Unlock()

			r.updateRegionCache(r.regions[regionIdx], true)
			return value, true, nil
		}

		if err != nil {
			lastErr = err
			r.updateRegionCache(r.regions[regionIdx], false)
			continue
		}
	}

	if lastErr != nil {
		return nil, false, lastErr
	}
	return nil, false, nil
}

func (r *RegionFailoverSource) updateRegionCache(region string, healthy bool) {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	if time.Since(r.lastCacheAt) > r.cacheTTL {
		r.regionCache = make(map[string]bool)
	}

	r.regionCache[region] = healthy
	r.lastCacheAt = time.Now()
}

// GetCurrentRegion returns the region that most recently served a lookup, or the first region
// if none has yet.
func (r *RegionFailoverSource) GetCurrentRegion() string {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()

	if r.currentRegion < len(r.regions) {
		return r.regions[r.currentRegion]
	}
	return ""
}

// GetHealthyRegions returns the regions currently recorded as healthy, in no particular order.
// A region appears only once it has been tried, and the record is cleared once it goes stale.
func (r *RegionFailoverSource) GetHealthyRegions() []string {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()

	healthy := make([]string, 0)
	for region, isHealthy := range r.regionCache {
		if isHealthy {
			healthy = append(healthy, region)
		}
	}
	return healthy
}

// FromAWSSecretsManagerMultiRegion reads secretName from the first responsive region, failing
// over between regions in the order given.
func FromAWSSecretsManagerMultiRegion(secretName string, regions []string) confkit.Source {
	src, err := NewRegionFailoverSource(regions, func(region string) confkit.Source {
		return FromAWSSecretsManagerWithRegion(secretName, region)
	})
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}

// FromAWSSSMParameterStoreMultiRegion reads parameters under pathPrefix from the first
// responsive region, with a five-minute cache per region and a trailing slash appended to the
// prefix if absent.
func FromAWSSSMParameterStoreMultiRegion(pathPrefix string, regions []string) confkit.Source {
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix += "/"
	}

	src, err := NewRegionFailoverSource(regions, func(region string) confkit.Source {
		ssmSrc, err := NewAWSSSMSourceWithRegion(pathPrefix, 5*time.Minute, region)
		if err != nil {
			return confkit.NewErrorSource(err)
		}
		return ssmSrc
	})
	if err != nil {
		return confkit.NewErrorSource(err)
	}
	return src
}
