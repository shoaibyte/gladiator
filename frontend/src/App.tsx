import { useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { NotebooksPage } from './pages/NotebooksPage'
import { NotebookEditorPage } from './pages/NotebookEditorPage'
import { ExplorePage } from './pages/ExplorePage'
import { ProtectedRoute } from './components/common/ProtectedRoute'
import { useAuthStore } from './store/authStore'

function App() {
  const loadUser = useAuthStore((s) => s.loadUser)
  useEffect(() => { loadUser() }, [loadUser])

  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/explore" element={<ExplorePage />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/notebooks" element={<NotebooksPage />} />
        <Route path="/notebooks/:id" element={<NotebookEditorPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
