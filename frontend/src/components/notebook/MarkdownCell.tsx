import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { Trash2, Edit3 } from 'lucide-react'
import type { Cell } from '../../types'

interface MarkdownCellProps {
  cell: Cell
  isActive: boolean
  onDelete: (id: string) => void
  onUpdate: (id: string, content: string) => void
  onClick: () => void
  isDark?: boolean
}

export function MarkdownCell({ cell, isActive, onDelete, onUpdate, onClick, isDark }: MarkdownCellProps) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(cell.content)

  const handleBlur = () => {
    setEditing(false)
    if (draft !== cell.content) {
      onUpdate(cell.id, draft)
    }
  }

  return (
    <div
      onClick={onClick}
      className={`rounded-2xl border-2 transition-all duration-200 overflow-hidden shadow-lg ${
        isActive
          ? 'border-indigo-400/80 dark:border-bronze-400/80 ring-2 ring-indigo-400/30 dark:ring-bronze-400/30 bg-white/25 dark:bg-slate-800/40'
          : 'border-white/25 dark:border-white/20 hover:border-white/40 dark:hover:border-white/30 bg-white/20 dark:bg-slate-800/30 hover:bg-white/25 dark:hover:bg-slate-800/40'
      } backdrop-blur-md`}
    >
      <div className="flex items-center justify-between px-4 py-2.5 border-b border-white/15 dark:border-white/10">
        <span className="text-[10px] font-bold uppercase tracking-widest text-amber-300/90 dark:text-amber-200/80">
          markdown
        </span>
        <div className="flex items-center gap-0.5">
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); setEditing((e) => !e); }}
            className="p-2 rounded-lg text-white/70 hover:bg-indigo-500/40 dark:hover:bg-bronze-500/40 hover:text-white transition-colors"
          >
            <Edit3 size={16} />
          </button>
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); onDelete(cell.id); }}
            className="p-2 rounded-lg text-white/50 hover:bg-red-500/30 hover:text-red-300 transition-colors"
          >
            <Trash2 size={16} />
          </button>
        </div>
      </div>
      {editing ? (
        <textarea
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={handleBlur}
          className="w-full p-4 min-h-[80px] font-mono text-sm bg-[#2E3440]/60 border-none focus:ring-0 resize-y text-[#D8DEE9] placeholder-[#4C566A] focus:outline-none"
          placeholder="Markdown..."
          autoFocus
        />
      ) : (
        <div
          className="p-4 font-mono text-sm text-[#D8DEE9] max-w-none prose prose-invert prose-sm
            prose-headings:font-mono prose-headings:font-semibold prose-headings:text-[#ECEFF4]
            prose-p:font-mono prose-p:text-[#D8DEE9] prose-p:my-2
            prose-li:font-mono prose-li:text-[#D8DEE9]
            prose-strong:font-mono prose-strong:text-[#ECEFF4] prose-strong:font-semibold
            prose-code:font-mono prose-code:text-sm prose-code:bg-[#3B4252] prose-code:text-[#88C0D0] prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:before:content-none prose-code:after:content-none
            prose-pre:font-mono prose-pre:text-sm prose-pre:bg-[#2E3440] prose-pre:text-[#D8DEE9] prose-pre:border prose-pre:border-[#3B4252]
            prose-a:text-[#88C0D0] prose-a:no-underline hover:prose-a:underline
            prose-blockquote:border-l-[#4C566A] prose-blockquote:text-[#4C566A] prose-blockquote:italic"
          onDoubleClick={() => setEditing(true)}
        >
          <ReactMarkdown>{cell.content || '*No content*'}</ReactMarkdown>
        </div>
      )}
    </div>
  )
}
