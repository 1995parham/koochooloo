import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api, atLeast, errMessage, type Role, type User } from './api'

const ROLES: readonly Role[] = ['user', 'admin', 'superadmin']

export function UsersPanel({ user }: { user: User }) {
  const [users, setUsers] = useState<User[]>([])
  const [error, setError] = useState<string | null>(null)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('user')

  const isSuper = atLeast(user.role, 'superadmin')

  const load = useCallback(async () => {
    try {
      setUsers(await api.listUsers())
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
      await api.createUser(username, password, role)
      setUsername('')
      setPassword('')
      setRole('user')
      await load()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  async function changeRole(id: number, next: Role) {
    try {
      await api.setRole(id, next)
      await load()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  async function remove(id: number, name: string) {
    if (!window.confirm(`Delete user ${name}?`)) return
    try {
      await api.deleteUser(id)
      await load()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  return (
    <section className="card">
      <h2>Users</h2>
      {error && <div className="error">{error}</div>}
      {isSuper && (
        <form className="row" onSubmit={create}>
          <input placeholder="username" value={username} onChange={(e) => setUsername(e.target.value)} required />
          <input
            placeholder="password (min 8)"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
          <select value={role} onChange={(e) => setRole(e.target.value as Role)}>
            {ROLES.map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
          <button type="submit">Add user</button>
        </form>
      )}
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Username</th>
            <th>Role</th>
            <th>Provider</th>
            {isSuper && <th />}
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id}>
              <td>{u.id}</td>
              <td>{u.username}</td>
              <td>
                {isSuper && u.id !== user.id ? (
                  <select value={u.role} onChange={(e) => void changeRole(u.id, e.target.value as Role)}>
                    {ROLES.map((r) => (
                      <option key={r} value={r}>
                        {r}
                      </option>
                    ))}
                  </select>
                ) : (
                  <span className={`badge role-${u.role}`}>{u.role}</span>
                )}
              </td>
              <td>{u.provider}</td>
              {isSuper && (
                <td>
                  {u.id !== user.id && (
                    <button className="danger" onClick={() => void remove(u.id, u.username)}>
                      delete
                    </button>
                  )}
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}
