import { useEffect, useState, type FormEvent } from 'react'
import { api, errMessage, setToken, type User } from './api'

interface Props {
  flash: string | null
  onLoggedIn: (user: User) => void
}

export function Login({ flash, onLoggedIn }: Props) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(flash)
  const [oidc, setOidc] = useState<{ enabled: boolean; url: string } | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api
      .authInfo()
      .then((info) => setOidc({ enabled: info.oidc_enabled, url: info.oidc_login_url }))
      .catch(() => setOidc(null))
  }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    setBusy(true)
    setError(null)
    try {
      const { token, user } = await api.login(username, password)
      setToken(token)
      onLoggedIn(user)
    } catch (err) {
      setError(errMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="center">
      <form className="card login" onSubmit={submit}>
        <h1>koochooloo</h1>
        <p className="muted">admin panel</p>
        {error && <div className="error">{error}</div>}
        <input
          placeholder="username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoFocus
        />
        <input
          placeholder="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button type="submit" disabled={busy}>
          {busy ? '…' : 'Sign in'}
        </button>
        {oidc?.enabled && (
          <>
            <div className="or">or</div>
            <a className="btn secondary" href={oidc.url}>
              Sign in with SSO
            </a>
          </>
        )}
      </form>
    </div>
  )
}
