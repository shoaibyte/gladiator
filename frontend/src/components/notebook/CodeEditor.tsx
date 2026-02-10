import { useCallback } from 'react'
import Editor from '@monaco-editor/react'

/** Polar Night theme (Nord-inspired): dark blue-grey background, clear syntax colors */
const POLAR_NIGHT_THEME = 'polar-night'

function definePolarNightTheme(monaco: { editor: { defineTheme: (name: string, theme: Record<string, unknown>) => void } }) {
  try {
    monaco.editor.defineTheme(POLAR_NIGHT_THEME, {
      base: 'vs-dark',
      inherit: true,
      rules: [
        { token: 'keyword', foreground: '81A1C1', fontStyle: 'bold' },
        { token: 'keyword.control', foreground: '81A1C1', fontStyle: 'bold' },
        { token: 'keyword.control.go', foreground: '81A1C1', fontStyle: 'bold' },
        { token: 'string', foreground: 'A3BE8C' },
        { token: 'string.quoted.double.go', foreground: 'A3BE8C' },
        { token: 'string.escape', foreground: '8FBCBB' },
        { token: 'comment', foreground: '4C566A', fontStyle: 'italic' },
        { token: 'comment.line.double-slash.go', foreground: '4C566A', fontStyle: 'italic' },
        { token: 'number', foreground: 'B48EAD' },
        { token: 'type', foreground: '8FBCBB' },
        { token: 'type.identifier', foreground: '88C0D0' },
        { token: 'identifier', foreground: 'D8DEE9' },
        { token: 'operator', foreground: '81A1C1' },
        { token: 'delimiter', foreground: 'D8DEE9' },
        { token: 'predefined', foreground: '88C0D0' },
        { token: 'function', foreground: '88C0D0' },
        { token: 'function.call.go', foreground: '88C0D0' },
        { token: 'variable', foreground: 'D8DEE9' },
        { token: 'constant', foreground: 'D08770' },
      ],
      colors: {
        'editor.background': '#2E3440',
        'editor.foreground': '#D8DEE9',
        'editorLineNumber.foreground': '#4C566A',
        'editorLineNumber.activeForeground': '#8FBCBB',
        'editor.selectionBackground': '#434C5E80',
        'editorCursor.foreground': '#D8DEE9',
        'editorIndentGuide.background': '#3B4252',
        'editorIndentGuide.activeBackground': '#4C566A',
      },
    })
  } catch {
    // theme may already be defined
  }
}

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
  const effectiveTheme = theme === 'vs-dark' ? POLAR_NIGHT_THEME : theme

  const handleBeforeMount = useCallback((monaco: { editor: { defineTheme: (name: string, theme: Record<string, unknown>) => void } }) => {
    definePolarNightTheme(monaco)
  }, [])

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
        beforeMount={handleBeforeMount}
        onMount={handleEditorMount}
        options={{
          minimap: { enabled: false },
          lineNumbers: 'on',
          fontSize: 14,
          tabSize: 4,
          wordWrap: 'on',
          readOnly: !!readOnly,
        }}
        theme={effectiveTheme}
      />
    </div>
  )
}
