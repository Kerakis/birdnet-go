package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tphakala/birdnet-go/internal/classifier"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/detection"
)

func TestCalculateMinDetectionsForLevel(t *testing.T) {
	t.Parallel()
	// overlap 2.4: segment 0.6, maxDet 10. level3=0.5->5, level5=0.7->7, level0->1.
	assert.Equal(t, 1, calculateMinDetectionsForLevel(0, 2.4))
	assert.Equal(t, 5, calculateMinDetectionsForLevel(3, 2.4))
	assert.Equal(t, 7, calculateMinDetectionsForLevel(5, 2.4))
}

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

func TestEffectiveMinDetections(t *testing.T) {
	t.Parallel()

	// Global: level 3, overlap 2.4 => minDetections 5 (see settingsForBirdMinDetections).
	base := settingsForBirdMinDetections(t, 5)
	base.Realtime.Species.Config = map[string]conf.SpeciesConfig{
		"strict bird":  {FilterLevel: new(5)}, // overlap 2.4 => 7
		"lenient bird": {FilterLevel: new(0)}, // => 1
		"no override":  {Threshold: 0.8},      // inherits global => 5
	}

	assert.Equal(t, 5, effectiveMinDetections(base, birdItem("Unknown", "Unknown sp")),
		"unconfigured species inherits global")
	assert.Equal(t, 5, effectiveMinDetections(base, birdItem("No Override", "no override")),
		"species with no FilterLevel inherits global")
	assert.Equal(t, 7, effectiveMinDetections(base, birdItem("Strict Bird", "strict bird")),
		"strict override raises the count")
	assert.Equal(t, 1, effectiveMinDetections(base, birdItem("Lenient Bird", "lenient bird")),
		"lenient override lowers the count")

	// Bat items ignore the per-species override entirely.
	batItem := birdItem("Strict Bird", "strict bird")
	batItem.BestModelID = classifier.RegistryIDBat
	assert.Equal(t, calculateBatMinDetections(base), effectiveMinDetections(base, batItem),
		"bat items ignore per-species FilterLevel")
}
