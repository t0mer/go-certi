const BASE = '/api/v1'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    credentials: 'include',
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export interface FQDN {
  id: string; fqdn: string; include_subdomains: boolean; enabled: boolean
  notifications_enabled: boolean; schedule_id: string | null; channel_ids: string[]
  created_at: string; updated_at: string
}
export interface Certificate {
  id: string; fqdn_id: string; serial: string; issuer_ca: string; subject_cn: string
  sans: string[]; not_before: string; not_after: string; discovered_at: string; source: string
}
export interface CertificateListResponse { items: Certificate[]; total: number; page: number; page_size: number }
export interface Channel {
  id: string; name: string; type: 'shoutrrr' | 'greenapi' | 'waweb'; config: string
  enabled: boolean; created_at: string; updated_at: string
}
export interface Schedule {
  id: string; name: string; cron_expr: string; is_default: boolean; enabled: boolean
  created_at: string; updated_at: string
}
export interface Settings {
  auth_enabled: boolean; username: string | null; api_token_protection_enabled: boolean
  theme: 'light' | 'dark' | 'system'; sslmate_api_key: string; default_schedule_id: string | null
}

export const api = {
  fqdns: {
    list: () => request<FQDN[]>('/fqdns'),
    get: (id: string) => request<FQDN>(`/fqdns/${id}`),
    create: (body: Partial<FQDN> & { fqdn: string }) => request<FQDN>('/fqdns', { method: 'POST', body: JSON.stringify(body) }),
    update: (id: string, body: Partial<FQDN>) => request<FQDN>(`/fqdns/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (id: string) => request<void>(`/fqdns/${id}`, { method: 'DELETE' }),
    scan: (id: string) => request<void>(`/fqdns/${id}/scan`, { method: 'POST' }),
  },
  certificates: {
    list: (params?: { fqdn?: string; page?: number; page_size?: number }) => {
      const q = new URLSearchParams()
      if (params?.fqdn) q.set('fqdn', params.fqdn)
      if (params?.page) q.set('page', String(params.page))
      if (params?.page_size) q.set('page_size', String(params.page_size))
      return request<CertificateListResponse>(`/certificates?${q}`)
    },
    get: (id: string) => request<Certificate>(`/certificates/${id}`),
    cas: () => request<string[]>('/certificates/cas'),
  },
  channels: {
    list: () => request<Channel[]>('/channels'),
    get: (id: string) => request<Channel>(`/channels/${id}`),
    create: (body: { name: string; type: string; config: Record<string, string>; enabled?: boolean }) =>
      request<Channel>('/channels', { method: 'POST', body: JSON.stringify(body) }),
    update: (id: string, body: { name: string; type: string; config: Record<string, string>; enabled: boolean }) =>
      request<Channel>(`/channels/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (id: string) => request<void>(`/channels/${id}`, { method: 'DELETE' }),
    test: (id: string) => request<{ status: string }>(`/channels/${id}/test`, { method: 'POST' }),
  },
  schedules: {
    list: () => request<Schedule[]>('/schedules'),
    create: (body: Partial<Schedule>) => request<Schedule>('/schedules', { method: 'POST', body: JSON.stringify(body) }),
    update: (id: string, body: Partial<Schedule>) => request<Schedule>(`/schedules/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
    delete: (id: string) => request<void>(`/schedules/${id}`, { method: 'DELETE' }),
  },
  settings: {
    get: () => request<Settings>('/settings'),
    update: (body: Partial<Settings> & { password?: string }) =>
      request<Settings>('/settings', { method: 'PUT', body: JSON.stringify(body) }),
    rotateToken: () => request<{ token: string }>('/settings/api-token/rotate', { method: 'POST' }),
  },
  auth: {
    login: (username: string, password: string) =>
      request<{ token: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
    logout: () => request<void>('/auth/logout', { method: 'POST' }),
    me: () => request<{ username: string }>('/auth/me'),
  },
}
