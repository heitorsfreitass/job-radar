import type { Filters } from '../types'

interface Props {
  filters: Filters
  onChange: (filters: Filters) => void
}

export default function FilterBar({ filters, onChange }: Props) {
  const set = <K extends keyof Filters>(key: K, value: Filters[K]) =>
    onChange({ ...filters, [key]: value })

  return (
    <div className="filter-bar">
      <input
        type="text"
        placeholder="Search title or description…"
        value={filters.keyword}
        onChange={(e) => set('keyword', e.target.value)}
        className="filter-keyword"
      />
      <input
        type="text"
        placeholder="Country"
        value={filters.country}
        onChange={(e) => set('country', e.target.value)}
      />
      <select value={filters.workplace} onChange={(e) => set('workplace', e.target.value)}>
        <option value="">Any workplace</option>
        <option value="remote">Remote</option>
        <option value="onsite">Onsite</option>
        <option value="hybrid">Hybrid</option>
      </select>
      <select value={filters.seniority} onChange={(e) => set('seniority', e.target.value)}>
        <option value="">Any seniority</option>
        <option value="junior">Junior</option>
        <option value="mid">Mid</option>
        <option value="senior">Senior</option>
        <option value="lead">Lead</option>
      </select>
      <input
        type="text"
        placeholder="Tag"
        value={filters.tag}
        onChange={(e) => set('tag', e.target.value)}
      />
    </div>
  )
}
