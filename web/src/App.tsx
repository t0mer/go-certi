import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import Dashboard from '@/pages/Dashboard'
import FQDNsPage from '@/pages/FQDNs'
import Certificates from '@/pages/Certificates'
import Channels from '@/pages/Channels'
import Schedules from '@/pages/Schedules'
import SettingsPage from '@/pages/Settings'
import Login from '@/pages/Login'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="fqdns" element={<FQDNsPage />} />
          <Route path="certificates" element={<Certificates />} />
          <Route path="channels" element={<Channels />} />
          <Route path="schedules" element={<Schedules />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
