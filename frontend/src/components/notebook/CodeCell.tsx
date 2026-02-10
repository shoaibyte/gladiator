import { useState } from 'react'
import { Play, Trash2, ChevronDown, ChevronUp } from 'lucide-react'
import { CodeEditor } from './CodeEditor'
import type { Cell } from '../../types'

interface CodeCellProps {
  cell: Cell
  isActive: boolean
  onExecute: (id: string) => void
  onDelete: (id: string) => void
  onUpdate: (id: string, content: string) => void
  onClick: () => void
  isDark?: boolean
}

export function CodeCell({ cell, isActive, onExecute, onDelete, onUpdate, onClick, isDark }: CodeCellProps) {
  const [showOutput, setShowOutput] = useState(true)

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
        <div className="flex items-center gap-3">
          <span className="text-[10px] font-bold uppercase tracking-widest text-indigo-300 dark:text-bronze-300/90">
            code
          </span>
          {cell.executed_at && (
            <span className="text-[10px] text-white/60 bg-white/10 dark:bg-slate-700/50 px-2 py-0.5 rounded-md">
              ran
            </span>
          )}
        </div>
        <div className="flex items-center gap-0.5">
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); onExecute(cell.id); }}
            className="p-2 rounded-lg text-white/70 hover:bg-indigo-500/40 dark:hover:bg-bronze-500/40 hover:text-white transition-colors"
            title="Execute (Ctrl+Enter)"
          >
            <Play size={16} />
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
      <div className="bg-slate-900/40 dark:bg-slate-950/50 border-t border-white/5">
<CodeEditor
        value={cell.content}
        onChange={(v) => onUpdate(cell.id, v)}
        onExecute={() => onExecute(cell.id)}
        readOnly={false}
        theme="vs-dark"
        height={Math.max(120, (cell.content.split('\n').length + 1) * 20)}
      />
      </div>
      {(cell.output != null && cell.output !== '') && (
        <div className="border-t border-white/15 dark:border-white/10">
          <button
            type="button"
            onClick={() => setShowOutput(!showOutput)}
            className="w-full flex items-center justify-between px-4 py-2.5 text-[10px] font-bold uppercase tracking-widest text-indigo-300/90 dark:text-bronze-300/90 hover:bg-white/10 transition-colors"
          >
            Output
            {showOutput ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          </button>
          {showOutput && (
            <pre className="p-4 font-mono text-sm bg-slate-900/90 dark:bg-slate-950 text-slate-100 overflow-x-auto whitespace-pre-wrap border-t border-indigo-500/20 dark:border-bronze-500/20">
              {cell.output}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}
