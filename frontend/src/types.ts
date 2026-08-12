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
