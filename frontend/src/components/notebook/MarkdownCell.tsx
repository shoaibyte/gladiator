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
      className={`rounded-2xl border-2 transition-all ${
        isActive
          ? 'border-indigo-500 dark:border-bronze-500 ring-2 ring-indigo-500/20 dark:ring-bronze-500/20'
          : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'
      } bg-white dark:bg-slate-800 overflow-hidden`}
    >
      <div className="flex items-center justify-between px-4 py-2 border-b border-slate-100 dark:border-slate-700">
        <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">markdown</span>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); setEditing((e) => !e); }}
            className="p-2 rounded-lg text-slate-400 hover:text-indigo-600 dark:hover:text-bronze-400"
          >
            <Edit3 size={16} />
          </button>
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); onDelete(cell.id); }}
            className="p-2 rounded-lg text-slate-400 hover:text-red-600 dark:hover:text-red-400"
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
          className="w-full p-4 min-h-[80px] font-mono text-sm bg-transparent border-none focus:ring-0 resize-y text-slate-800 dark:text-slate-200"
          placeholder="Markdown..."
          autoFocus
        />
      ) : (
        <div
          className="p-4 prose dark:prose-invert prose-slate max-w-none dark:prose-headings:text-white dark:prose-p:text-slate-300"
          onDoubleClick={() => setEditing(true)}
        >
          <ReactMarkdown>{cell.content || '*No content*'}</ReactMarkdown>
        </div>
      )}
    </div>
  )
}
