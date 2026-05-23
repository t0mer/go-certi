import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useTheme } from '@/lib/theme'

export default function SettingsPage() {
  const qc = useQueryClient()
  const { setTheme } = useTheme()
  const [newToken, setNewToken] = useState<string | null>(null)
  const [password, setPassword] = useState('')
  const [saved, setSaved] = useState(false)

  const { data: settings, isLoading } = useQuery({ queryKey: ['settings'], queryFn: api.settings.get })

  const [form, setForm] = useState({
    auth_enabled: false,
    username: '',
    api_token_protection_enabled: false,
    theme: 'system' as 'light' | 'dark' | 'system',
    sslmate_api_key: '',
    default_schedule_id: null as string | null,
  })

  useEffect(() => {
    if (settings) {
      setForm({
        auth_enabled: settings.auth_enabled,
        username: settings.username ?? '',
        api_token_protection_enabled: settings.api_token_protection_enabled,
        theme: settings.theme,
        sslmate_api_key: settings.sslmate_api_key,
        default_schedule_id: settings.default_schedule_id,
      })
    }
  }, [settings])

  const updateMutation = useMutation({
    mutationFn: () => api.settings.update({
      ...form,
      username: form.username || undefined,
      password: password || undefined,
    }),
    onSuccess: (updated) => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      setTheme(updated.theme)
      setPassword('')
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  const rotateMutation = useMutation({
    mutationFn: api.settings.rotateToken,
    onSuccess: (data) => setNewToken(data.token),
  })

  if (isLoading) return <div className="text-muted-foreground py-8 text-center">Loading...</div>

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold">Settings</h1>

      <Card>
        <CardHeader><CardTitle>Appearance</CardTitle></CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground mb-2">Theme</p>
          <div className="flex gap-2">
            {(['light', 'dark', 'system'] as const).map((t) => (
              <Button key={t} size="sm" variant={form.theme === t ? 'default' : 'outline'}
                onClick={() => setForm({ ...form, theme: t })}>
                {t}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>sslmate API Key</CardTitle></CardHeader>
        <CardContent className="space-y-2">
          <Input
            type="password"
            placeholder="Enter sslmate API key..."
            value={form.sslmate_api_key}
            onChange={(e) => setForm({ ...form, sslmate_api_key: e.target.value })}
          />
          <p className="text-xs text-muted-foreground">Leave empty to use crt.sh (fallback, no key required).</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>UI Authentication</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" checked={form.auth_enabled}
              onChange={(e) => setForm({ ...form, auth_enabled: e.target.checked })} />
            <span className="text-sm">Require login to access the UI</span>
          </label>
          {form.auth_enabled && (
            <>
              <Input placeholder="Username" value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })} />
              <Input type="password" placeholder="New password (leave blank to keep current)"
                value={password} onChange={(e) => setPassword(e.target.value)} />
            </>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>API Token Protection</CardTitle></CardHeader>
        <CardContent className="space-y-3">
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" checked={form.api_token_protection_enabled}
              onChange={(e) => setForm({ ...form, api_token_protection_enabled: e.target.checked })} />
            <span className="text-sm">Require Bearer token for all API requests</span>
          </label>
          {form.api_token_protection_enabled && (
            <div className="space-y-2">
              <Button variant="outline" size="sm" onClick={() => { setNewToken(null); rotateMutation.mutate() }}
                disabled={rotateMutation.isPending}>
                {rotateMutation.isPending ? 'Generating...' : 'Generate new token'}
              </Button>
              {newToken && (
                <div className="bg-muted rounded-md p-3">
                  <p className="text-xs text-muted-foreground mb-1 font-medium">New token — copy now, won't be shown again:</p>
                  <code className="text-xs break-all select-all">{newToken}</code>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="flex items-center gap-3">
        <Button onClick={() => updateMutation.mutate()} disabled={updateMutation.isPending}>
          {updateMutation.isPending ? 'Saving...' : 'Save Settings'}
        </Button>
        {saved && <span className="text-sm text-green-600 dark:text-green-400">Saved</span>}
        {updateMutation.isError && <span className="text-sm text-destructive">{(updateMutation.error as Error).message}</span>}
      </div>
    </div>
  )
}
