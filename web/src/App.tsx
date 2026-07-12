import { useCallback, useEffect, useState } from 'react'
import { api, clearToken, getToken, setToken, type User } from './api'
import { Dashboard } from './Dashboard'
import { Login } from './Login'

export function App() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [flash, setFlash] = useState<string | null>(null)

  // Pick up a token (or error) handed back in the URL fragment by the OIDC
  // callback redirect, then scrub it from the address bar.
  useEffect(() => {
    const hash = window.location.hash
    if (hash.startsWith('#token=')) {
      setToken(decodeURIComponent(hash.slice('#token='.length)))
      window.history.replaceState(null, '', '/admin/')
    } else if (hash.startsWith('#error=')) {
      setFlash(`SSO login failed: ${decodeURIComponent(hash.slice('#error='.length))}`)
      window.history.replaceState(null, '', '/admin/')
    }
  }, [])

  const refresh = useCallback(async () => {
    if (!getToken()) {
      setUser(null)
      setLoading(false)
      return
    }
    try {
      setUser(await api.me())
    } catch {
      clearToken()
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  if (loading) return <div className="center muted">loading…</div>
  if (!user) return <Login flash={flash} onLoggedIn={setUser} />
  return (
    <Dashboard
      user={user}
      onLogout={() => {
        clearToken()
        setUser(null)
      }}
    />
  )
}
