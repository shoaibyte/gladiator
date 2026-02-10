import { useCallback } from 'react'
import Editor from '@monaco-editor/react'

interface CodeEditorProps {
  value: string
  onChange: (value: string) => void
  onExecute?: () => void
  onSave?: () => void
  readOnly?: boolean
  theme?: 'light' | 'vs-dark'
  height?: number
}

export function CodeEditor({ value, onChange, onExecute, onSave, readOnly, theme = 'vs-dark', height = 200 }: CodeEditorProps) {
  const handleEditorMount = useCallback((editor: unknown, monaco: unknown) => {
    const e = editor as { addAction: (opts: { id: string; label: string; keybindings: number[]; run: () => void }) => void }
    const m = monaco as { KeyMod: { CtrlCmd: number }; KeyCode: { Enter: number; KeyS: number } }
    if (!e?.addAction || !m) return
    e.addAction({
      id: 'execute',
      label: 'Execute',
      keybindings: [m.KeyMod.CtrlCmd | m.KeyCode.Enter],
      run: () => onExecute?.(),
    })
    e.addAction({
      id: 'save',
      label: 'Save',
      keybindings: [m.KeyMod.CtrlCmd | m.KeyCode.KeyS],
      run: () => onSave?.(),
    })
  }, [onExecute, onSave])

  return (
    <div style={{ minHeight: Math.min(Math.max(100, height), 600) }}>
      <Editor
        height={height}
        defaultLanguage="go"
        value={value}
        onChange={(v) => onChange(v ?? '')}
        onMount={handleEditorMount}
        options={{
          minimap: { enabled: false },
          lineNumbers: 'on',
          fontSize: 14,
          tabSize: 4,
          wordWrap: 'on',
          readOnly: !!readOnly,
        }}
        theme={theme}
      />
    </div>
  )
}
