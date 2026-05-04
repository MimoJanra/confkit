package confkit

import (
	"fmt"
	"sync"
	"time"
)

type RegionFailoverSource struct {
	regions       []string
	sources       []Source
	currentRegion int
	cacheMutex    sync.RWMutex
	regionCache   map[string]bool
	cacheTTL      time.Duration
	lastCacheAt   time.Time
}

func NewRegionFailoverSource(regions []string, sourceFactory func(region string) Source) (*RegionFailoverSource, error) {
	if len(regions) == 0 {
		return nil, fmt.Errorf("must provide at least one region")
	}

	sources := make([]Source, len(regions))
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

func (r *RegionFailoverSource) Name() string {
	return "multiregion"
}

func (r *RegionFailoverSource) Lookup(field *FieldInfo) (any, bool, error) {
	r.cacheMutex.RLock()
	startRegion := r.currentRegion
	r.cacheMutex.RUnlock()

	for i := 0; i < len(r.regions); i++ {
		regionIdx := (startRegion + i) % len(r.regions)
		source := r.sources[regionIdx]

		value, ok, err := source.Lookup(field)
		if err == nil && ok {
			r.cacheMutex.Lock()
			r.currentRegion = regionIdx
			r.cacheMutex.Unlock()

			r.updateRegionCache(r.regions[regionIdx], true)
			return value, true, nil
		}

		if err != nil {
			r.updateRegionCache(r.regions[regionIdx], false)
			continue
		}
	}

	return "", false, fmt.Errorf("all regions exhausted for field %s", field.Path)
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

func (r *RegionFailoverSource) GetCurrentRegion() string {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()

	if r.currentRegion < len(r.regions) {
		return r.regions[r.currentRegion]
	}
	return ""
}

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

func FromAWSSecretsManagerMultiRegion(secretName string, regions []string) Source {
	src, err := NewRegionFailoverSource(regions, func(region string) Source {
		return FromAWSSecretsManagerWithRegion(secretName, region)
	})
	if err != nil {
		return &errorSource{err: err}
	}
	return src
}

func FromAWSSSMParameterStoreMultiRegion(pathPrefix string, regions []string) Source {
	src, err := NewRegionFailoverSource(regions, func(region string) Source {
		ssmSrc, _ := NewAWSSSMSource(pathPrefix, 5*time.Minute)
		return ssmSrc
	})
	if err != nil {
		return &errorSource{err: err}
	}
	return src
}
