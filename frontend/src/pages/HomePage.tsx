import { Link } from 'react-router-dom'
import { BookOpen, Sword } from 'lucide-react'

export function HomePage() {
  return (
    <div className="min-h-screen min-h-[100dvh] relative text-slate-900 dark:text-white overflow-hidden">
      {/* Viking mascot background - responsive for all viewport sizes */}
      <img
        src="/viking-landing.png"
        alt=""
        className="absolute inset-0 w-full h-full object-cover object-[center_40%] sm:object-center"
        decoding="async"
        fetchPriority="high"
      />
      <div className="absolute inset-0 bg-slate-900/50 sm:bg-slate-900/60 dark:bg-slate-950/70" aria-hidden />

      <div className="relative z-10 min-h-screen min-h-[100dvh] flex flex-col">
        <nav className="border-b border-white/10 px-4 py-3 sm:px-6 sm:py-4">
          <div className="max-w-7xl mx-auto flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="h-9 w-9 sm:h-10 sm:w-10 shrink-0 rounded-xl bg-indigo-600 dark:bg-bronze-600 flex items-center justify-center">
                <Sword className="h-5 w-5 sm:h-6 sm:w-6 text-white dark:text-black" />
              </div>
              <span className="text-lg sm:text-xl font-bold text-white truncate">GoViking</span>
            </div>
            <div className="flex gap-2 sm:gap-4 shrink-0">
              <Link
                to="/login"
                className="text-xs sm:text-sm font-medium text-white/90 hover:text-white py-2 px-2 sm:px-0"
              >
                Sign in
              </Link>
              <Link
                to="/register"
                className="rounded-xl px-3 py-2 sm:px-4 sm:py-2 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-semibold text-xs sm:text-sm hover:opacity-90 whitespace-nowrap"
              >
                Enter The Arena
              </Link>
            </div>
          </div>
        </nav>
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 sm:py-16 md:py-20 lg:py-24 text-center flex-1 flex flex-col justify-center">
          <h1 className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-black tracking-tight mb-3 sm:mb-4 text-white drop-shadow-lg">
            The Magical <span className="text-indigo-300 dark:text-bronze-400">Go Arena</span>
          </h1>
          <p className="text-base sm:text-lg md:text-xl text-white/90 max-w-2xl mx-auto mb-8 sm:mb-10 md:mb-12 drop-shadow px-1">
            Code that feels like magic. Execute your Go spells instantly, collaborate with your guild, and build legendary software.
          </p>
          <Link
            to="/register"
            className="inline-flex items-center justify-center gap-2 rounded-2xl px-6 py-3 sm:px-8 sm:py-4 bg-indigo-600 dark:bg-bronze-600 text-white dark:text-black font-bold text-base sm:text-lg hover:opacity-90 shadow-xl w-full sm:w-auto max-w-xs sm:max-w-none mx-auto"
          >
            <BookOpen className="h-5 w-5 shrink-0" /> Join the Guild
          </Link>
        </main>
      </div>
    </div>
  )
}
