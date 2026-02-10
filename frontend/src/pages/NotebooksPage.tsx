import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '../services/api'
import { useAuthStore } from '../store/authStore'
import { Plus, BookOpen } from 'lucide-react'

interface NotebookMeta {
  id: string
  title: string
  description?: string | null
  is_public: boolean
  created_at: string
  updated_at: string
  cell_count: number
}

export function NotebooksPage() {
  const logout = useAuthStore((s) => s.logout)
  const [search, setSearch] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: ['notebooks', search],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (search) params.set('search', search)
      const { data } = await api.get<{ notebooks: NotebookMeta[]; pagination: { page: number; limit: number; total: number; total_pages: number } }>('/notebooks?' + params)
      return data
    },
  })

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 dark:border-slate-800 px-6 py-4 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold text-slate-900 dark:text-white">Gladiator</Link>
        <div className="flex items-center gap-4">
          <input
            type="search"
            placeholder="Search notebooks..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 px-4 py-2 w-64 text-sm"
          />
          <Link
            to="/notebooks/new"
            className="inline-flex items-center gap-2 rounded-xl px-4 py-2 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold text-sm"
          >
            <Plus className="h-4 w-4" /> New Notebook
          </Link>
          <button onClick={logout} className="text-sm text-slate-500 hover:text-slate-700 dark:hover:text-slate-300">
            Logout
          </button>
        </div>
      </header>
      <main className="max-w-6xl mx-auto px-6 py-12">
        <h1 className="text-3xl font-bold mb-8 text-slate-900 dark:text-white">My Notebooks</h1>
        {isLoading ? (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-40 rounded-2xl bg-slate-200 dark:bg-slate-700 animate-pulse" />
            ))}
          </div>
        ) : !data?.notebooks?.length ? (
          <div className="rounded-2xl border-2 border-dashed border-slate-200 dark:border-slate-700 p-16 text-center text-slate-500">
            <BookOpen className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>No notebooks yet. Create one to enter the arena.</p>
            <Link to="/notebooks/new" className="mt-4 inline-block rounded-xl px-6 py-2 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold">
              New Notebook
            </Link>
          </div>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {data.notebooks.map((nb) => (
              <Link
                key={nb.id}
                to={`/notebooks/${nb.id}`}
                className="block rounded-2xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-6 hover:shadow-lg transition-shadow"
              >
                <h2 className="font-bold text-lg text-slate-900 dark:text-white truncate">{nb.title}</h2>
                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">{nb.description || 'No description'}</p>
                <p className="text-xs text-slate-400 mt-2">{nb.cell_count} cells · Updated {new Date(nb.updated_at).toLocaleDateString()}</p>
                {nb.is_public && <span className="inline-block mt-2 text-xs px-2 py-0.5 rounded bg-indigo-100 dark:bg-bronze-900/30 text-indigo-700 dark:text-bronze-400">Public</span>}
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  )
}
