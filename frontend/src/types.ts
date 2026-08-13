export interface Job {
  ID: string
  Title: string
  CompanyName: string
  Description: string
  URL: string
  Source: string
  SourceJobID: string
  Country: string
  Workplace: string
  Employment: string
  Seniority: string
  Tags: string[]
  PublishedAt: string
  IngestedAt: string
}

export interface SearchMeta {
  page: number
  page_size: number
  total: number
}

export interface SearchResponse {
  data: Job[]
  meta: SearchMeta
}

export interface Filters {
  keyword: string
  country: string
  workplace: string
  seniority: string
  tag: string
}

export const EMPTY_FILTERS: Filters = {
  keyword: '',
  country: '',
  workplace: '',
  seniority: '',
  tag: '',
}

export interface AuthResponse {
  token: string
  email: string
}

// Matches the JSON shape of Go's domain.Preferences (capitalized field
// names, no json tags) — distinct from the frontend's own lowercase
// Filters shape.
export interface PreferencesDTO {
  Country: string
  Workplace: string
  Seniority: string
  Tag: string
  Keyword: string
}

export interface MeResponse {
  email: string
  preferences: PreferencesDTO
}

export function preferencesToFilters(p: PreferencesDTO): Filters {
  return {
    country: p.Country,
    workplace: p.Workplace,
    seniority: p.Seniority,
    tag: p.Tag,
    keyword: p.Keyword,
  }
}

export function filtersToPreferences(f: Filters): PreferencesDTO {
  return {
    Country: f.country,
    Workplace: f.workplace,
    Seniority: f.seniority,
    Tag: f.tag,
    Keyword: f.keyword,
  }
}
