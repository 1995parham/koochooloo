import { useEffect, useState } from 'react'
import { api, atLeast, type User, type Version } from './api'
import { UrlsPanel } from './UrlsPanel'
import { UsersPanel } from './UsersPanel'

interface Props {
  user: User
  onLogout: () => void
}

export function Dashboard({ user, onLogout }: Props) {
  const [version, setVersion] = useState<Version | null>(null)

  useEffect(() => {
    api
      .version()
      .then(setVersion)
      .catch(() => setVersion(null))
  }, [])

  return (
    <div className="app">
      <header>
        <div className="brand">
          <strong>koochooloo</strong> <span className="muted">admin</span>
        </div>
        <div className="spacer" />
        <span className="badge">{user.username}</span>
        <span className={`badge role-${user.role}`}>{user.role}</span>
        <button type="button" className="secondary" onClick={onLogout}>
          Logout
        </button>
      </header>
      <main>
        <UrlsPanel user={user} />
        {atLeast(user.role, 'admin') && <UsersPanel user={user} />}
      </main>
      {version && <Footer version={version} />}
    </div>
  )
}

// Footer shows the running build. The server only sends the commit and its
// timestamp to admins, so those parts simply stay absent for everyone else.
function Footer({ version }: { version: Version }) {
  const commit = version.revision?.slice(0, 7)

  return (
    <footer className="version muted">
      <span>{version.version}</span>
      {commit && (
        <>
          <span aria-hidden="true">·</span>
          <code title={version.revision}>
            {commit}
            {version.dirty && '-dirty'}
          </code>
        </>
      )}
      {version.last_commit && (
        <>
          <span aria-hidden="true">·</span>
          <time dateTime={version.last_commit}>
            {new Date(version.last_commit).toLocaleString()}
          </time>
        </>
      )}
    </footer>
  )
}
