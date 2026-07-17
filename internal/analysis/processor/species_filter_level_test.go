package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tphakala/birdnet-go/internal/classifier"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/detection"
)

// TestSpeciesConfigHasFilterOverride verifies detection of a per-species
// FilterLevel override, including that level 0 counts as a real override.
func TestSpeciesConfigHasFilterOverride(t *testing.T) {
	t.Parallel()
	assert.False(t, speciesConfigHasFilterOverride(nil), "nil map has no override")
	assert.False(t, speciesConfigHasFilterOverride(map[string]conf.SpeciesConfig{
		"a": {Threshold: 0.5}, "b": {Interval: 30},
	}), "configs without FilterLevel have no override")
	assert.True(t, speciesConfigHasFilterOverride(map[string]conf.SpeciesConfig{
		"a": {Threshold: 0.5}, "b": {FilterLevel: new(0)},
	}), "a FilterLevel (even 0) counts as an override")
}

// TestCalculateMinDetectionsForLevel checks the pure level+overlap
// min-detections calculation against known values.
func TestCalculateMinDetectionsForLevel(t *testing.T) {
	t.Parallel()
	// overlap 2.4: segment 0.6, maxDet 10. level3=0.5->5, level5=0.7->7, level0->1.
	assert.Equal(t, 1, calculateMinDetectionsForLevel(0, 2.4))
	assert.Equal(t, 5, calculateMinDetectionsForLevel(3, 2.4))
	assert.Equal(t, 7, calculateMinDetectionsForLevel(5, 2.4))
}

// birdItem builds a bird-model PendingDetection for the given common and
// scientific names, used to exercise per-species lookup in the tests.
func birdItem(common, scientific string) *PendingDetection {
	return &PendingDetection{
		Detection: Detections{
			Result: detection.Result{
				Species: detection.Species{CommonName: common, ScientificName: scientific},
			},
		},
		BestModelID: "BirdNET_V2.4",
	}
}

// TestEffectiveMinDetections verifies the per-item required-count resolution:
// inherit, stricter and more-permissive overrides, the no-override fast path,
// and that bat items ignore per-species overrides.
func TestEffectiveMinDetections(t *testing.T) {
	t.Parallel()

	// Global: level 3, overlap 2.4 => minDetections 5 (see settingsForBirdMinDetections).
	base := settingsForBirdMinDetections(t, 5)
	base.Realtime.Species.Config = map[string]conf.SpeciesConfig{
		"strict bird":  {FilterLevel: new(5)}, // overlap 2.4 => 7
		"lenient bird": {FilterLevel: new(0)}, // => 1
		"no override":  {Threshold: 0.8},      // inherits global => 5
	}
	has := speciesConfigHasFilterOverride(base.Realtime.Species.Config)
	assert.True(t, has, "base config has overrides")

	assert.Equal(t, 5, effectiveMinDetections(base, birdItem("Unknown", "Unknown sp"), has),
		"unconfigured species inherits global")
	assert.Equal(t, 5, effectiveMinDetections(base, birdItem("No Override", "no override"), has),
		"species with no FilterLevel inherits global")
	assert.Equal(t, 7, effectiveMinDetections(base, birdItem("Strict Bird", "strict bird"), has),
		"strict override raises the count")
	assert.Equal(t, 1, effectiveMinDetections(base, birdItem("Lenient Bird", "lenient bird"), has),
		"lenient override lowers the count")

	// Fast path: when no overrides exist (hasOverride=false), the per-item lookup is
	// skipped and every bird species inherits the global count.
	assert.Equal(t, 5, effectiveMinDetections(base, birdItem("Strict Bird", "strict bird"), false),
		"hasOverride=false skips the species lookup")

	// Bat items ignore the per-species override entirely.
	batItem := birdItem("Strict Bird", "strict bird")
	batItem.BestModelID = classifier.RegistryIDBat
	assert.Equal(t, calculateBatMinDetections(base), effectiveMinDetections(base, batItem, has),
		"bat items ignore per-species FilterLevel")
}
