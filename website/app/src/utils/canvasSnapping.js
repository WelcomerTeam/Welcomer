function findClosestSnap(position, anchors, targets, tolerance) {
  let closestSnap = null;

  for (const anchor of anchors) {
    for (const target of targets) {
      const delta = target - anchor;
      const distance = Math.abs(delta);

      if (distance <= tolerance && (!closestSnap || distance < closestSnap.distance)) {
        closestSnap = {
          position: position + delta,
          guide: target,
          distance,
        };
      }
    }
  }

  return closestSnap
    ? { position: closestSnap.position, guide: closestSnap.guide }
    : null;
}

export function getCanvasSnapTargets(canvasSize, borderWidth, layers, objectIndex, coordinate) {
  const canvasStart = -borderWidth;
  const canvasEnd = canvasSize - borderWidth;
  const intervals = layers
    .map((layer, index) => {
      if (index === objectIndex) return null;

      const start = layer.position?.[coordinate] || 0;
      const size = layer.dimensions?.[coordinate] || 0;
      return size > 0 ? { start, end: start + size } : null;
    })
    .filter(Boolean)
    .sort((a, b) => a.start - b.start);
  const targets = {
    edges: [canvasStart, canvasEnd],
    centers: [(canvasStart + canvasEnd) / 2],
    gaps: [],
  };

  let rightmostInterval = null;
  intervals.forEach((interval) => {
    targets.edges.push(interval.start, interval.end);
    targets.centers.push((interval.start + interval.end) / 2);

    if (!rightmostInterval) {
      rightmostInterval = interval;
      return;
    }

    if (interval.start >= rightmostInterval.end) {
      targets.gaps.push({
        start: rightmostInterval.end,
        end: interval.start,
      });
      rightmostInterval = interval;
    } else if (interval.end > rightmostInterval.end) {
      rightmostInterval = interval;
    }
  });

  return targets;
}

function findClosestGapSnap(position, size, gaps, tolerance) {
  let closestSnap = null;

  for (const gap of gaps) {
    const gapSize = gap.end - gap.start;
    if (size > gapSize) continue;

    const snappedPosition = gap.start + (gapSize - size) / 2;
    const distance = Math.abs(snappedPosition - position);

    if (distance <= tolerance && (!closestSnap || distance < closestSnap.distance)) {
      closestSnap = {
        position: snappedPosition,
        guide: gap.start + gapSize / 2,
        distance,
      };
    }
  }

  return closestSnap
    ? { position: closestSnap.position, guide: closestSnap.guide }
    : null;
}

export function getSnappedPosition(position, size, targets, tolerance, preferredEdge = null) {
  const left = position;
  const center = position + size / 2;
  const right = position + size;

  if (preferredEdge === "start" || preferredEdge === "end") {
    const preferredAnchor = preferredEdge === "end" ? right : left;
    const preferredSnap = findClosestSnap(
      position,
      [preferredAnchor],
      targets.edges,
      tolerance
    );
    if (preferredSnap) return preferredSnap;
  }

  const edgeAnchors = preferredEdge === "end"
    ? [left]
    : preferredEdge === "start"
      ? [right]
      : [left, right];
  const edgeSnap = findClosestSnap(position, edgeAnchors, targets.edges, tolerance);
  if (edgeSnap) return edgeSnap;

  const gapSnap = findClosestGapSnap(position, size, targets.gaps || [], tolerance);
  if (gapSnap) return gapSnap;

  const centerSnap = findClosestSnap(position, [center], targets.centers, tolerance);
  return centerSnap || { position, guide: null };
}
