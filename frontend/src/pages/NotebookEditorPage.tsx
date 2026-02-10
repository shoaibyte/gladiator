import { useEffect, useState, useCallback, useRef } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, BookOpen, Plus, Code, FileText } from 'lucide-react'
import { api } from '../services/api'
import type { Notebook, Cell } from '../types'
import { CodeCell } from '../components/notebook/CodeCell'
import { MarkdownCell } from '../components/notebook/MarkdownCell'
import { useDebounce } from '../hooks/useDebounce'

function normalizeCell(c: unknown): Cell {
  const o = c as Record<string, unknown>
  const id = typeof o?.id === 'string' ? o.id : crypto.randomUUID()
  const type = (o?.type === 'code' || o?.type === 'markdown') ? o.type : 'code'
  const content = typeof o?.content === 'string' ? o.content : ''
  const output = o?.output != null ? String(o.output) : null
  const executed_at = o?.executed_at != null ? String(o.executed_at) : null
  const order = typeof o?.order === 'number' ? o.order : 0
  return { id, type: type as 'code' | 'markdown', content, output, executed_at, order }
}

export function NotebookEditorPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeCellId, setActiveCellId] = useState<string | null>(null)
  const [title, setTitle] = useState('')
  const [cells, setCells] = useState<Cell[]>([])
  const [saveStatus, setSaveStatus] = useState<'saved' | 'saving' | 'error'>('saved')
  const isDark = document.documentElement.classList.contains('dark')
  const lastSavedRef = useRef<string>('')

  const { data: notebook, isLoading } = useQuery({
    queryKey: ['notebook', id],
    queryFn: async () => {
      const { data } = await api.get<Notebook>(`/notebooks/${id}`)
      return data
    },
    enabled: !!id && id !== 'new',
  })

  useEffect(() => {
    if (id === 'new') {
      api.post<Notebook>('/notebooks', { title: 'Untitled' })
        .then(({ data }) => navigate(`/notebooks/${data.id}`, { replace: true }))
        .catch(() => {})
    }
  }, [id, navigate])

  useEffect(() => {
    if (notebook) {
      setTitle(notebook.title)
      const raw = (notebook.content?.cells ?? []) as unknown[]
      const next = raw.map(normalizeCell).sort((a, b) => a.order - b.order)
      setCells(next)
      lastSavedRef.current = JSON.stringify(next)
    }
  }, [notebook])

  const debouncedCells = useDebounce(cells, 2000)
  const patchMutation = useMutation({
    mutationFn: (payload: { title?: string; content?: { cells: Cell[] } }) =>
      api.patch(`/notebooks/${id}`, payload),
    onMutate: () => setSaveStatus('saving'),
    onSuccess: () => { setSaveStatus('saved'); queryClient.invalidateQueries({ queryKey: ['notebook', id] }) },
    onError: () => setSaveStatus('error'),
  })

  useEffect(() => {
    if (!id || id === 'new') return
    const key = JSON.stringify(debouncedCells)
    if (key === lastSavedRef.current) return
    lastSavedRef.current = key
    patchMutation.mutate({ content: { cells: debouncedCells } })
  }, [debouncedCells, id])

  const addCell = useCallback((type: 'code' | 'markdown') => {
    const newCell: Cell = {
      id: crypto.randomUUID(),
      type,
      content: type === 'code' ? 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("Hello")\n}' : '',
      output: null,
      executed_at: null,
      order: cells.length,
    }
    setCells((prev) => [...prev, newCell])
    setActiveCellId(newCell.id)
  }, [cells.length])

  const updateCell = useCallback((cellId: string, content: string) => {
    setCells((prev) => prev.map((c) => (c.id === cellId ? { ...c, content } : c)))
  }, [])

  const deleteCell = useCallback((cellId: string) => {
    setCells((prev) => prev.filter((c) => c.id !== cellId).map((c, i) => ({ ...c, order: i })))
    setActiveCellId((cur) => (cur === cellId ? null : cur))
  }, [])

  const executeCell = useCallback(async (cellId: string) => {
    const cell = cells.find((c) => c.id === cellId)
    if (!cell || cell.type !== 'code' || !id) return
    try {
      const { data } = await api.post<{ stdout: string; stderr: string; status: string }>(`/notebooks/${id}/execute`, { cell_id: cellId, code: cell.content })
      const out = data.stdout ?? ''
      const err = data.stderr ?? ''
      setCells((prev) =>
        prev.map((c) =>
          c.id === cellId ? { ...c, output: out + (err ? '\n' + err : ''), executed_at: new Date().toISOString() } : c
        )
      )
      queryClient.invalidateQueries({ queryKey: ['notebook', id] })
    } catch {
      setCells((prev) =>
        prev.map((c) => (c.id === cellId ? { ...c, output: 'Execution failed', executed_at: new Date().toISOString() } : c))
      )
    }
  }, [cells, id, queryClient])

  if (id === 'new') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 dark:border-bronze-500" />
      </div>
    )
  }

  if (isLoading || !notebook) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-900">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 dark:border-bronze-500" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-slate-50 dark:bg-slate-900 flex flex-col">
      <header className="border-b border-slate-200 dark:border-slate-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/notebooks" className="p-2 rounded-xl hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-600 dark:bg-bronze-600 text-white">
            <BookOpen className="h-5 w-5" />
          </div>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={() => patchMutation.mutate({ title })}
            className="text-lg font-bold bg-transparent border-none focus:ring-0 text-slate-900 dark:text-white w-64"
          />
          <span className="text-xs text-slate-500 dark:text-slate-400">
            {saveStatus === 'saving' && 'Saving...'}
            {saveStatus === 'saved' && 'Saved'}
            {saveStatus === 'error' && 'Error saving'}
          </span>
        </div>
      </header>
      <main className="flex-1 p-8 max-w-4xl mx-auto w-full space-y-6">
        {cells.map((cell) =>
          cell.type === 'code' ? (
            <CodeCell
              key={cell.id}
              cell={cell}
              isActive={activeCellId === cell.id}
              onExecute={executeCell}
              onDelete={deleteCell}
              onUpdate={updateCell}
              onClick={() => setActiveCellId(cell.id)}
              isDark={isDark}
            />
          ) : (
            <MarkdownCell
              key={cell.id}
              cell={cell}
              isActive={activeCellId === cell.id}
              onDelete={deleteCell}
              onUpdate={updateCell}
              onClick={() => setActiveCellId(cell.id)}
              isDark={isDark}
            />
          )
        )}
        <div className="flex gap-4 pt-8">
          <button
            type="button"
            onClick={() => addCell('code')}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-indigo-100 dark:bg-bronze-500/20 text-indigo-700 dark:text-bronze-400 font-semibold text-sm"
          >
            <Code size={18} /> Add code cell
          </button>
          <button
            type="button"
            onClick={() => addCell('markdown')}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 font-semibold text-sm"
          >
            <FileText size={18} /> Add markdown
          </button>
        </div>
      </main>
    </div>
  )
}

