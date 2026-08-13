import type { AuthResponse, Filters, Job, MeResponse, PreferencesDTO, SearchResponse } from './types'

const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export async function searchJobs(filters: Filters, page: number, pageSize: number): Promise<SearchResponse> {
  const params = new URLSearchParams()
  if (filters.keyword) params.set('keyword', filters.keyword)
  if (filters.country) params.set('country', filters.country)
  if (filters.workplace) params.set('workplace', filters.workplace)
  if (filters.seniority) params.set('seniority', filters.seniority)
  if (filters.tag) params.set('tag', filters.tag)
  params.set('page', String(page))
  params.set('page_size', String(pageSize))

  const res = await fetch(`${API_BASE}/jobs?${params.toString()}`)
  if (!res.ok) {
    throw new Error(`search failed: ${res.status}`)
  }
  return res.json()
}

export async function getJob(id: string): Promise<Job> {
  const res = await fetch(`${API_BASE}/jobs/${id}`)
  if (!res.ok) {
    throw new Error(`get job failed: ${res.status}`)
  }
  return res.json()
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? `request failed: ${res.status}`)
  }
  return res.json()
}

export function register(email: string, password: string): Promise<AuthResponse> {
  return postJSON('/auth/register', { email, password })
}

export function login(email: string, password: string): Promise<AuthResponse> {
  return postJSON('/auth/login', { email, password })
}

export async function getMe(token: string): Promise<MeResponse> {
  const res = await fetch(`${API_BASE}/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) {
    throw new Error(`get me failed: ${res.status}`)
  }
  return res.json()
}

export async function savePreferences(token: string, prefs: PreferencesDTO): Promise<void> {
  const res = await fetch(`${API_BASE}/me/preferences`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(prefs),
  })
  if (!res.ok) {
    throw new Error(`save preferences failed: ${res.status}`)
  }
}
