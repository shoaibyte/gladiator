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
    <div className="min-h-screen min-h-[100dvh] relative overflow-hidden">
      <img
        src="/viking-notebooks.png"
        alt=""
        className="absolute inset-0 w-full h-full object-cover object-center"
        decoding="async"
      />
      <div className="absolute inset-0 bg-slate-900/50 dark:bg-slate-950/60" aria-hidden />
      <div className="relative z-10 min-h-screen min-h-[100dvh] flex flex-col">
        <header className="border-b border-white/10 px-4 sm:px-6 py-4 flex items-center justify-between gap-2 flex-wrap">
          <Link to="/" className="text-lg sm:text-xl font-bold text-white">GoViking</Link>
          <div className="flex items-center gap-2 sm:gap-4">
            <input
              type="search"
              placeholder="Search notebooks..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="rounded-xl border border-white/20 bg-white/20 dark:bg-slate-800/40 backdrop-blur-sm text-white placeholder-white/70 px-4 py-2 w-40 sm:w-64 text-sm focus:ring-2 focus:ring-indigo-400 outline-none"
            />
            <Link
              to="/notebooks/new"
              className="inline-flex items-center gap-2 rounded-xl px-3 py-2 sm:px-4 sm:py-2 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold text-sm"
            >
              <Plus className="h-4 w-4" /> New Notebook
            </Link>
            <button onClick={logout} className="text-sm text-white/80 hover:text-white">
              Logout
            </button>
          </div>
        </header>
        <main className="max-w-6xl mx-auto px-4 sm:px-6 py-8 sm:py-12 flex-1">
          <h1 className="text-2xl sm:text-3xl font-bold mb-6 sm:mb-8 text-white drop-shadow-lg">My GoBooks</h1>
          {isLoading ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5 lg:gap-6">
              {[1, 2, 3].map((i) => (
                <div key={i} className="min-h-[200px] sm:min-h-[220px] lg:min-h-[240px] rounded-2xl bg-white/10 dark:bg-slate-800/30 backdrop-blur-sm animate-pulse" />
              ))}
            </div>
          ) : !data?.notebooks?.length ? (
            <div className="rounded-2xl border-2 border-dashed border-white/30 bg-white/10 dark:bg-slate-800/20 backdrop-blur-sm p-12 sm:p-16 text-center text-white/90">
              <BookOpen className="h-12 w-12 mx-auto mb-4 opacity-80" />
              <p>No notebooks yet. Create one to enter the arena.</p>
              <Link to="/notebooks/new" className="mt-4 inline-block rounded-xl px-6 py-2 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold">
                New Notebook
              </Link>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-5 lg:gap-6">
              {data.notebooks.map((nb) => (
                <Link
                  key={nb.id}
                  to={`/notebooks/${nb.id}`}
                  className="group flex flex-col rounded-2xl border border-white/20 bg-white/15 dark:bg-slate-800/25 backdrop-blur-md p-6 sm:p-7 lg:p-8 min-h-[200px] sm:min-h-[220px] lg:min-h-[240px] hover:bg-white/25 dark:hover:bg-slate-800/40 hover:shadow-xl hover:scale-[1.02] transition-all duration-200 shadow-lg"
                >
                  <div className="flex justify-center mb-4 sm:mb-5">
                    <div className="flex h-12 w-12 sm:h-14 sm:w-14 lg:h-16 lg:w-16 items-center justify-center rounded-xl bg-indigo-500/30 dark:bg-bronze-500/30 text-indigo-200 dark:text-bronze-200 group-hover:bg-indigo-500/40 dark:group-hover:bg-bronze-500/40 transition-colors">
                      <BookOpen className="h-6 w-6 sm:h-7 sm:w-7 lg:h-8 lg:w-8" />
                    </div>
                  </div>
                  <h2 className="font-bold text-lg sm:text-xl text-white truncate drop-shadow">{nb.title}</h2>
                  <p className="text-sm sm:text-base text-white/80 mt-2 line-clamp-2 flex-1">{nb.description || 'No description'}</p>
                  <p className="text-xs sm:text-sm text-white/60 mt-3">{nb.cell_count} cells · Updated {new Date(nb.updated_at).toLocaleDateString()}</p>
                  {nb.is_public && <span className="inline-block mt-2 text-xs px-2 py-0.5 rounded bg-indigo-500/30 text-indigo-200 w-fit">Public</span>}
                </Link>
              ))}
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
