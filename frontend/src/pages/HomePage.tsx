import { Link } from 'react-router-dom'
import { BookOpen, Sword } from 'lucide-react'

export function HomePage() {
  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900 text-slate-900 dark:text-white">
      <nav className="border-b border-slate-200 dark:border-slate-800 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="h-10 w-10 rounded-xl bg-indigo-600 dark:bg-bronze-600 flex items-center justify-center">
              <Sword className="h-6 w-6 text-white dark:text-black" />
            </div>
            <span className="text-xl font-bold">Gladiator</span>
          </div>
          <div className="flex gap-4">
            <Link to="/login" className="text-sm font-medium text-slate-600 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-bronze-400">
              Sign in
            </Link>
            <Link
              to="/register"
              className="rounded-xl px-4 py-2 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold text-sm hover:opacity-90"
            >
              Enter The Arena
            </Link>
          </div>
        </div>
      </nav>
      <main className="max-w-7xl mx-auto px-6 py-24 text-center">
        <h1 className="text-5xl font-black tracking-tight mb-4">
          The Magical <span className="text-indigo-600 dark:text-bronze-500">Go Arena</span>
        </h1>
        <p className="text-xl text-slate-600 dark:text-slate-400 max-w-2xl mx-auto mb-12">
          Code that feels like magic. Execute your Go spells instantly, collaborate with your guild, and build legendary software.
        </p>
        <Link
          to="/register"
          className="inline-flex items-center gap-2 rounded-2xl px-8 py-4 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-bold text-lg hover:opacity-90"
        >
          <BookOpen className="h-5 w-5" /> Join the Guild
        </Link>
      </main>
    </div>
  )
}
