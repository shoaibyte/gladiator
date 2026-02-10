import { Link } from 'react-router-dom'
import { BookOpen } from 'lucide-react'

export function ExplorePage() {
  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900">
      <header className="border-b border-slate-200 dark:border-slate-800 px-6 py-4">
        <Link to="/" className="text-xl font-bold text-slate-900 dark:text-white">GoViking</Link>
      </header>
      <main className="max-w-6xl mx-auto px-6 py-12">
        <h1 className="text-3xl font-bold mb-8 text-slate-900 dark:text-white">Explore Public Notebooks</h1>
        <p className="text-slate-500 dark:text-slate-400">Public notebooks will appear here.</p>
      </main>
    </div>
  )
}
