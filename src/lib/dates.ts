// Date helpers operating on ISO strings (YYYY-MM-DD). Kept out of .svelte files
// so Date arithmetic stays in plain TS (the svelte/prefer-svelte-reactivity rule
// flags raw Date/Map in components) and is reusable across screens.

function at(isoStr: string): Date {
  return new Date(isoStr + 'T00:00:00')
}

export function iso(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function todayISO(): string {
  return iso(new Date())
}

export function addDays(isoStr: string, n: number): string {
  const d = at(isoStr)
  d.setDate(d.getDate() + n)
  return iso(d)
}

export function fmtLong(isoStr: string): string {
  return at(isoStr).toLocaleDateString('pt-BR', {
    weekday: 'long',
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

export interface WeekDay {
  iso: string
  label: string
  num: number
}

/** Seven days starting at the anchor. */
export function weekDays(anchorISO: string): WeekDay[] {
  const out: WeekDay[] = []
  for (let i = 0; i < 7; i++) {
    const isoStr = addDays(anchorISO, i)
    const d = at(isoStr)
    out.push({
      iso: isoStr,
      label: d.toLocaleDateString('pt-BR', { weekday: 'short' }),
      num: d.getDate()
    })
  }
  return out
}

export interface MonthCell {
  iso: string
  num: number
  dim: boolean
}

/** A 6×7 Monday-first month grid containing the anchor's month. */
export function monthGrid(anchorISO: string): MonthCell[] {
  const a = at(anchorISO)
  const first = new Date(a.getFullYear(), a.getMonth(), 1)
  const start = new Date(first)
  start.setDate(start.getDate() - ((first.getDay() + 6) % 7))
  const out: MonthCell[] = []
  for (let i = 0; i < 42; i++) {
    const d = new Date(start)
    d.setDate(d.getDate() + i)
    out.push({ iso: iso(d), num: d.getDate(), dim: d.getMonth() !== a.getMonth() })
  }
  return out
}
