export interface User {
  id: string
  email: string
  name: string
  avatar_url?: string | null
  created_at: string
}

export interface Cell {
  id: string
  type: 'code' | 'markdown'
  content: string
  output: string | null
  executed_at: string | null
  order: number
}

export interface NotebookContent {
  cells: Cell[]
}

export interface Notebook {
  id: string
  owner_id: string
  title: string
  description: string | null
  content: NotebookContent
  is_public: boolean
  created_at: string
  updated_at: string
  last_executed_at: string | null
  execution_count: number
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface WebSocketMessage {
  type: string
  notebook_id: string
  user_id: string
  user_name: string
  payload: unknown
  timestamp: number
}
