import type { Job } from './types'

const PREFIX = 'job-radar:'

function loadJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(PREFIX + key)
    return raw ? (JSON.parse(raw) as T) : fallback
  } catch {
    return fallback
  }
}

function saveJSON<T>(key: string, value: T): void {
  localStorage.setItem(PREFIX + key, JSON.stringify(value))
}

export function loadHiddenIds(): Set<string> {
  return new Set(loadJSON<string[]>('hidden', []))
}

export function toggleHidden(current: Set<string>, id: string): Set<string> {
  const next = new Set(current)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  saveJSON('hidden', [...next])
  return next
}

// Saved jobs store the full Job, not just the id, so the "Saved" view can
// render them even when they've scrolled off the current search results
// (or the API's ingestion later drops/updates them).
export function loadSavedJobs(): Record<string, Job> {
  return loadJSON<Record<string, Job>>('saved', {})
}

export function toggleSaved(current: Record<string, Job>, job: Job): Record<string, Job> {
  const next = { ...current }
  if (next[job.ID]) {
    delete next[job.ID]
  } else {
    next[job.ID] = job
  }
  saveJSON('saved', next)
  return next
}

export { loadJSON, saveJSON }
