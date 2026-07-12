// api.ts — typed client for the koochooloo admin API. Response shapes are
// declared once as zod schemas; the TypeScript types are inferred from them and
// every response is validated at runtime, so the network boundary is checked
// rather than merely asserted. The JWT is kept in localStorage and attached as
// a bearer token on every request.

import { z } from 'zod'

const TOKEN_KEY = 'koochooloo_token'
const BASE = '/admin/api'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}
export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}
export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export const RoleSchema = z.enum(['user', 'admin', 'superadmin'])
export type Role = z.infer<typeof RoleSchema>

const RANK: Record<Role, number> = { user: 1, admin: 2, superadmin: 3 }

export function atLeast(role: Role, min: Role): boolean {
  return RANK[role] >= RANK[min]
}

export const UserSchema = z.object({
  id: z.number(),
  username: z.string(),
  role: RoleSchema,
  provider: z.string(),
  created_at: z.string(),
})
export type User = z.infer<typeof UserSchema>

export const UrlSchema = z.object({
  key: z.string(),
  url: z.string(),
  count: z.number(),
  expire_time: z.string().nullable(),
  owner_id: z.number().nullable(),
})
export type Url = z.infer<typeof UrlSchema>

export const AuthInfoSchema = z.object({
  oidc_enabled: z.boolean(),
  oidc_login_url: z.string(),
})
export type AuthInfo = z.infer<typeof AuthInfoSchema>

const TokenSchema = z.object({ token: z.string(), user: UserSchema })
const CreatedKeySchema = z.object({ key: z.string() })

/** Unauthorized is thrown when the server rejects the current token. */
export class Unauthorized extends Error {}

/** errMessage extracts a human-readable message from a caught value. */
export function errMessage(e: unknown): string {
  return e instanceof Error ? e.message : String(e)
}

async function request(path: string, opts: RequestInit = {}): Promise<Response> {
  const headers = new Headers(opts.headers)
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (opts.body !== undefined && opts.body !== null) headers.set('Content-Type', 'application/json')

  const res = await fetch(BASE + path, { ...opts, headers })
  if (res.status === 401) throw new Unauthorized('session expired')
  return res
}

async function fail(res: Response, fallback: string): Promise<never> {
  let message = fallback
  try {
    const body: unknown = await res.json()
    const parsed = z.object({ message: z.string() }).safeParse(body)
    if (parsed.success) message = parsed.data.message
  } catch {
    // ignore non-JSON error bodies
  }
  throw new Error(message)
}

/** parse validates a response body against a schema, or throws. */
async function parse<T>(res: Response, schema: z.ZodType<T>, fallback: string): Promise<T> {
  if (!res.ok) return fail(res, fallback)
  const body: unknown = await res.json()
  return schema.parse(body)
}

export const api = {
  authInfo(): Promise<AuthInfo> {
    return request('/auth/info').then((r) => parse(r, AuthInfoSchema, 'failed to load auth info'))
  },

  login(username: string, password: string): Promise<z.infer<typeof TokenSchema>> {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }).then((r) => parse(r, TokenSchema, 'login failed'))
  },

  me(): Promise<User> {
    return request('/auth/me').then((r) => parse(r, UserSchema, 'failed to load profile'))
  },

  listUrls(): Promise<Url[]> {
    return request('/urls').then((r) => parse(r, z.array(UrlSchema), 'failed to list urls'))
  },

  createUrl(url: string, name: string, expire: string | null): Promise<{ key: string }> {
    const body: { url: string; name?: string; expire?: string } = { url }
    if (name) body.name = name
    if (expire) body.expire = expire
    return request('/urls', { method: 'POST', body: JSON.stringify(body) }).then((r) =>
      parse(r, CreatedKeySchema, 'failed to create url'),
    )
  },

  async deleteUrl(key: string): Promise<void> {
    const res = await request(`/urls/${encodeURIComponent(key)}`, { method: 'DELETE' })
    if (!res.ok) await fail(res, 'failed to delete url')
  },

  listUsers(): Promise<User[]> {
    return request('/users').then((r) => parse(r, z.array(UserSchema), 'failed to list users'))
  },

  createUser(username: string, password: string, role: Role): Promise<User> {
    return request('/users', {
      method: 'POST',
      body: JSON.stringify({ username, password, role }),
    }).then((r) => parse(r, UserSchema, 'failed to create user'))
  },

  async setRole(id: number, role: Role): Promise<void> {
    const res = await request(`/users/${id}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    })
    if (!res.ok) await fail(res, 'failed to change role')
  },

  async deleteUser(id: number): Promise<void> {
    const res = await request(`/users/${id}`, { method: 'DELETE' })
    if (!res.ok) await fail(res, 'failed to delete user')
  },
}
