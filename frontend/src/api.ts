import type { Filters, Job, SearchResponse } from './types'

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
