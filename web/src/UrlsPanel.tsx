import { type FormEvent, useCallback, useEffect, useState } from 'react'
import { api, atLeast, errMessage, type Url, type User } from './api'

export function UrlsPanel({ user }: { user: User }) {
  const [urls, setUrls] = useState<Url[]>([])
  const [url, setUrl] = useState('')
  const [name, setName] = useState('')
  const [expire, setExpire] = useState('')
  const [error, setError] = useState<string | null>(null)

  const isAdmin = atLeast(user.role, 'admin')
  const origin = window.location.origin

  const load = useCallback(async () => {
    try {
      setUrls(await api.listUrls())
    } catch (e) {
      setError(errMessage(e))
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  async function create(event: FormEvent) {
    event.preventDefault()
    setError(null)
    try {
      await api.createUrl(url, name, expire ? new Date(expire).toISOString() : null)
      setUrl('')
      setName('')
      setExpire('')
      await load()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  async function remove(key: string) {
    if (!window.confirm(`Delete ${key}?`)) return
    try {
      await api.deleteUrl(key)
      await load()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  return (
    <section className="card">
      <h2>Short URLs {isAdmin && <span className="muted">· all users</span>}</h2>
      {error && <div className="error">{error}</div>}
      <form className="row" onSubmit={create}>
        <input
          placeholder="https://long.url/…"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          required
        />
        <input
          placeholder="custom name (optional)"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input type="datetime-local" value={expire} onChange={(e) => setExpire(e.target.value)} />
        <button type="submit">Create</button>
      </form>
      <table>
        <thead>
          <tr>
            <th>Short</th>
            <th>Target</th>
            <th>Hits</th>
            {isAdmin && <th>Owner</th>}
            <th />
          </tr>
        </thead>
        <tbody>
          {urls.map((u) => (
            <tr key={u.key}>
              <td className="nowrap">
                <a href={`/api/${u.key}`} target="_blank" rel="noreferrer">
                  {u.key}
                </a>
                <button
                  type="button"
                  className="link"
                  onClick={() => void navigator.clipboard.writeText(`${origin}/api/${u.key}`)}
                >
                  copy
                </button>
              </td>
              <td className="truncate">{u.url}</td>
              <td>{u.count}</td>
              {isAdmin && <td>{u.owner_id ?? '—'}</td>}
              <td>
                <button type="button" className="danger" onClick={() => void remove(u.key)}>
                  delete
                </button>
              </td>
            </tr>
          ))}
          {urls.length === 0 && (
            <tr>
              <td colSpan={isAdmin ? 5 : 4} className="muted">
                no short urls yet
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </section>
  )
}
