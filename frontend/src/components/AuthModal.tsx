import { useState } from 'react'
import { login, register } from '../api'
import type { AuthResponse } from '../types'

interface Props {
  onSuccess: (auth: AuthResponse) => void
  onClose: () => void
}

export default function AuthModal({ onSuccess, onClose }: Props) {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      const auth = mode === 'login' ? await login(email, password) : await register(email, password)
      onSuccess(auth)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal auth-modal" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose} aria-label="Close">
          ×
        </button>
        <h2>{mode === 'login' ? 'Log in' : 'Create account'}</h2>
        <p className="job-detail-company">
          {mode === 'login' ? "Don't have an account yet?" : 'Already have an account?'}{' '}
          <button className="link-btn" type="button" onClick={() => setMode(mode === 'login' ? 'register' : 'login')}>
            {mode === 'login' ? 'Sign up' : 'Log in'}
          </button>
        </p>

        <form onSubmit={handleSubmit} className="auth-form">
          <input
            type="email"
            placeholder="Email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
          />
          <input
            type="password"
            placeholder="Password (min. 8 characters)"
            required
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
          />
          {error && <p className="state-error auth-error">{error}</p>}
          <button type="submit" className="auth-submit" disabled={loading}>
            {loading ? 'Please wait…' : mode === 'login' ? 'Log in' : 'Sign up'}
          </button>
        </form>
      </div>
    </div>
  )
}
