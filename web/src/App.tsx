import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import FQDNsPage from '@/pages/FQDNs'
import Certificates from '@/pages/Certificates'

const Placeholder = ({ title }: { title: string }) => (
  <div className="text-center py-12">
    <h1 className="text-2xl font-bold mb-2">{title}</h1>
    <p className="text-muted-foreground">Coming soon...</p>
  </div>
)

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Placeholder title="Login" />} />
        <Route element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="fqdns" element={<FQDNsPage />} />
          <Route path="certificates" element={<Certificates />} />
          <Route path="channels" element={<Placeholder title="Channels" />} />
          <Route path="schedules" element={<Placeholder title="Schedules" />} />
          <Route path="settings" element={<Placeholder title="Settings" />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
