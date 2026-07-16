import { describe, it, expect } from 'vitest';
import { calculateMinDetections } from './filterLevels';

describe('calculateMinDetections', () => {
  it('returns 1 for level 0 regardless of overlap', () => {
    expect(calculateMinDetections(0, 2.8)).toBe(1);
  });
  it('matches backend counts at overlap 2.4', () => {
    expect(calculateMinDetections(3, 2.4)).toBe(5);
    expect(calculateMinDetections(5, 2.4)).toBe(7);
  });
  it('matches backend counts at documented overlaps', () => {
    expect(calculateMinDetections(4, 2.7)).toBe(12);
    expect(calculateMinDetections(5, 2.8)).toBe(21);
  });
});
