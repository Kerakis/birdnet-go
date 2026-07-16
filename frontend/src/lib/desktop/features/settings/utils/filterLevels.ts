/**
 * Shared bird false-positive filter level definitions.
 *
 * Single source of truth for the 0-5 bird filter levels, their badge styles,
 * the overlap-aware minimum-detections calculation (mirrors the Go backend in
 * internal/analysis/processor/processor.go), and the level description text.
 * Consumed by both the global filter control (AnalysisSettingsPage) and the
 * per-species override control (SpeciesConfigEditor).
 */
import type { FilterLevel } from '$lib/desktop/components/forms/FalsePositiveFilterControl.svelte';
import { safeArrayAccess } from '$lib/utils/security';
import { t } from '$lib/i18n';

export const BADGE_OFF = 'bg-black/5 dark:bg-white/5 text-[var(--color-base-content)]';
export const BADGE_SUCCESS = 'bg-[var(--color-success)] text-[var(--color-success-content)]';
export const BADGE_INFO = 'bg-[var(--color-info)] text-[var(--color-info-content)]';
export const BADGE_WARNING = 'bg-[var(--color-warning)] text-[var(--color-warning-content)]';
export const BADGE_ERROR = 'bg-[var(--color-error)] text-[var(--color-error-content)]';

/** Bird filter levels for the stepped slider (value + label key + badge). */
export const BIRD_FP_LEVELS: FilterLevel[] = [
  {
    value: 0,
    nameKey: 'settings.main.sections.falsePositiveFilter.levelNames.off',
    badgeClass: BADGE_OFF,
  },
  {
    value: 1,
    nameKey: 'settings.main.sections.falsePositiveFilter.levelNames.lenient',
    badgeClass: BADGE_SUCCESS,
  },
  {
    value: 2,
    nameKey: 'settings.main.sections.falsePositiveFilter.levelNames.moderate',
    badgeClass: BADGE_INFO,
  },
  {
    value: 3,
    nameKey: 'settings.main.sections.falsePositiveFilter.levelNames.balanced',
    badgeClass: BADGE_WARNING,
  },
  {
    value: 4,
    nameKey: 'settings.main.sections.falsePositiveFilter.levelNames.strict',
    badgeClass: BADGE_ERROR,
  },
  {
    value: 5,
    nameKey: 'settings.main.sections.falsePositiveFilter.levelNames.maximum',
    badgeClass: BADGE_ERROR,
  },
];

/** Per-level metadata: description key, minimum overlap, and threshold fraction. */
export const birdFilterLevelMeta = [
  {
    value: 0,
    descriptionKey: 'settings.main.sections.falsePositiveFilter.levels.off',
    minOverlap: 0.0,
    threshold: 0.0,
  },
  {
    value: 1,
    descriptionKey: 'settings.main.sections.falsePositiveFilter.levels.lenient',
    minOverlap: 2.0,
    threshold: 0.2,
  },
  {
    value: 2,
    descriptionKey: 'settings.main.sections.falsePositiveFilter.levels.moderate',
    minOverlap: 2.2,
    threshold: 0.3,
  },
  {
    value: 3,
    descriptionKey: 'settings.main.sections.falsePositiveFilter.levels.balanced',
    minOverlap: 2.4,
    threshold: 0.5,
  },
  {
    value: 4,
    descriptionKey: 'settings.main.sections.falsePositiveFilter.levels.strict',
    minOverlap: 2.7,
    threshold: 0.6,
  },
  {
    value: 5,
    descriptionKey: 'settings.main.sections.falsePositiveFilter.levels.maximum',
    minOverlap: 2.8,
    threshold: 0.7,
  },
];

// Constants matching backend: internal/analysis/processor/processor.go
const CHUNK_DURATION_SECONDS = 3.0;
const REFERENCE_WINDOW_SECONDS = 6.0;
const MIN_SEGMENT_LENGTH = 0.1;
const FLOAT_EPSILON = 1e-9;

/**
 * Computes the minimum number of confirmations required for a bird filter level
 * at the given overlap. Mirrors calculateMinDetectionsForLevel in the Go backend.
 */
export function calculateMinDetections(level: number, overlap: number): number {
  if (level === 0) return 1;
  const meta = safeArrayAccess(birdFilterLevelMeta, level);
  if (!meta) return 1;
  const segmentLength = Math.max(MIN_SEGMENT_LENGTH, CHUNK_DURATION_SECONDS - overlap);
  const maxDetectionsIn6s = REFERENCE_WINDOW_SECONDS / segmentLength;
  const required = maxDetectionsIn6s * meta.threshold - FLOAT_EPSILON;
  return Math.max(1, Math.ceil(required));
}

/**
 * Returns the localized description for a bird filter level, including the
 * effective confirmation count for the given overlap.
 */
export function getBirdFilterDescription(level: number, overlap: number): string {
  const meta = safeArrayAccess(birdFilterLevelMeta, level);
  if (!meta) return '';
  const baseDescription = t(meta.descriptionKey);
  if (level === 0) return baseDescription;
  return t('settings.main.sections.falsePositiveFilter.detectionCount', {
    count: calculateMinDetections(level, overlap).toString(),
    description: baseDescription,
  });
}
