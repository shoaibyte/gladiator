import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { api } from '../services/api'
import type { TokenPair } from '../types'

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const setTokens = useAuthStore((s) => s.setTokens)
  const setUser = useAuthStore((s) => s.setUser)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      const { data } = await api.post<TokenPair>('/auth/login', { email, password })
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      setTokens(data.access_token, data.refresh_token)
      setUser(data.user)
      window.location.href = '/notebooks'
    } catch (err: unknown) {
      setError((err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Login failed')
    }
  }

  return (
    <div className="min-h-screen min-h-[100dvh] relative flex items-center justify-center p-4 overflow-hidden">
      <img
        src="/viking-login.png"
        alt=""
        className="absolute inset-0 w-full h-full object-cover object-center"
        decoding="async"
      />
      <div className="absolute inset-0 bg-slate-900/55 dark:bg-slate-950/65" aria-hidden />
      <div className="relative z-10 w-full max-w-md rounded-2xl border border-white/25 bg-white/15 dark:bg-slate-900/20 backdrop-blur-md p-6 sm:p-8 shadow-2xl">
        <h1 className="text-2xl font-bold mb-6 text-center text-slate-900 dark:text-white drop-shadow-[0_1px_2px_rgba(0,0,0,0.8)]">Sign in to GoViking</h1>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-xl border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-900 px-4 py-3 focus:ring-2 focus:ring-indigo-500 dark:focus:ring-bronze-500 outline-none"
            required
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-xl border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-900 px-4 py-3 focus:ring-2 focus:ring-indigo-500 dark:focus:ring-bronze-500 outline-none"
            required
          />
          <button
            type="submit"
            className="w-full rounded-xl py-3 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold hover:opacity-90"
          >
            Sign in
          </button>
        </form>
        <p className="mt-4 text-center text-sm text-slate-700 dark:text-slate-300 drop-shadow-[0_1px_1px_rgba(0,0,0,0.6)]">
          No account? <Link to="/register" className="text-indigo-600 dark:text-bronze-400 font-medium">Register</Link>
        </p>
      </div>
    </div>
  )
}
