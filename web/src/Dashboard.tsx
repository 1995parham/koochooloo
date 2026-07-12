import { atLeast, type User } from './api'
import { UrlsPanel } from './UrlsPanel'
import { UsersPanel } from './UsersPanel'

interface Props {
  user: User
  onLogout: () => void
}

export function Dashboard({ user, onLogout }: Props) {
  return (
    <div className="app">
      <header>
        <div className="brand">
          <strong>koochooloo</strong> <span className="muted">admin</span>
        </div>
        <div className="spacer" />
        <span className="badge">{user.username}</span>
        <span className={`badge role-${user.role}`}>{user.role}</span>
        <button className="secondary" onClick={onLogout}>
          Logout
        </button>
      </header>
      <main>
        <UrlsPanel user={user} />
        {atLeast(user.role, 'admin') && <UsersPanel user={user} />}
      </main>
    </div>
  )
}
