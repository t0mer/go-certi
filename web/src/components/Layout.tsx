import { Link, useLocation, Outlet } from 'react-router-dom'
import { Shield, Globe, FileText, Bell, Clock, Settings, LayoutDashboard, Moon, Sun } from 'lucide-react'
import { useTheme } from '@/lib/theme'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/fqdns', label: 'FQDNs', icon: Globe },
  { to: '/certificates', label: 'Certificates', icon: FileText },
  { to: '/channels', label: 'Channels', icon: Bell },
  { to: '/schedules', label: 'Schedules', icon: Clock },
  { to: '/settings', label: 'Settings', icon: Settings },
]

export function Layout() {
  const location = useLocation()
  const { theme, setTheme } = useTheme()
  return (
    <div className="flex h-screen bg-background">
      <aside className="hidden md:flex w-56 flex-col border-r bg-card">
        <div className="flex h-14 items-center gap-2 border-b px-4">
          <Shield className="h-5 w-5 text-primary" />
          <span className="font-bold text-lg">go-certi</span>
        </div>
        <nav className="flex-1 p-2 space-y-1">
          {navItems.map(({ to, label, icon: Icon }) => (
            <Link key={to} to={to} className={cn('flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors', location.pathname === to ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground')}>
              <Icon className="h-4 w-4" />{label}
            </Link>
          ))}
        </nav>
        <div className="p-2 border-t">
          <Button variant="ghost" size="sm" className="w-full justify-start gap-2" onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
            {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            {theme === 'dark' ? 'Light mode' : 'Dark mode'}
          </Button>
        </div>
      </aside>
      <div className="flex flex-col flex-1 overflow-hidden">
        <header className="md:hidden flex h-14 items-center gap-2 border-b px-4">
          <Shield className="h-5 w-5 text-primary" />
          <span className="font-bold">go-certi</span>
        </header>
        <main className="flex-1 overflow-auto p-4 md:p-6"><Outlet /></main>
        <nav className="md:hidden flex border-t">
          {navItems.map(({ to, icon: Icon }) => (
            <Link key={to} to={to} className={cn('flex-1 flex flex-col items-center py-2 text-xs', location.pathname === to ? 'text-primary' : 'text-muted-foreground')}>
              <Icon className="h-5 w-5" />
            </Link>
          ))}
        </nav>
      </div>
    </div>
  )
}
