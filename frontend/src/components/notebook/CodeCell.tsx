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
      className={`rounded-2xl border-2 transition-all ${
        isActive
          ? 'border-indigo-500 dark:border-bronze-500 ring-2 ring-indigo-500/20 dark:ring-bronze-500/20'
          : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'
      } bg-white dark:bg-slate-800 overflow-hidden`}
    >
      <div className="flex items-center justify-between px-4 py-2 border-b border-slate-100 dark:border-slate-700">
        <div className="flex items-center gap-3">
          <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">code</span>
          {cell.executed_at && (
            <span className="text-xs text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded-full">
              {cell.executed_at}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); onExecute(cell.id); }}
            className="p-2 rounded-lg text-slate-600 dark:text-slate-400 hover:bg-indigo-100 dark:hover:bg-bronze-500/20 hover:text-indigo-600 dark:hover:text-bronze-400"
            title="Execute (Ctrl+Enter)"
          >
            <Play size={16} />
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
      <CodeEditor
        value={cell.content}
        onChange={(v) => onUpdate(cell.id, v)}
        onExecute={() => onExecute(cell.id)}
        readOnly={false}
        theme={isDark ? 'vs-dark' : 'light'}
        height={Math.max(120, (cell.content.split('\n').length + 1) * 20)}
      />
      {(cell.output != null && cell.output !== '') && (
        <div className="border-t border-slate-100 dark:border-slate-700">
          <button
            type="button"
            onClick={() => setShowOutput(!showOutput)}
            className="w-full flex items-center justify-between px-4 py-2 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider hover:bg-slate-50 dark:hover:bg-slate-700/50"
          >
            Output
            {showOutput ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          </button>
          {showOutput && (
            <pre className="p-4 font-mono text-sm bg-slate-900 text-slate-100 overflow-x-auto whitespace-pre-wrap border-t border-slate-700">
              {cell.output}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}
