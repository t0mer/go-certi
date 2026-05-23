import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Send, Bell, Pencil, X } from 'lucide-react'
import { api, type Channel } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const defaultConfigs: Record<string, Record<string, string>> = {
  shoutrrr: { url: '' },
  greenapi: { instance_id: '', api_token_instance: '', chat_id: '', api_url: 'https://api.green-api.com' },
  waweb: { base_url: '', phone: '', auth: '' },
}

const typeLabels: Record<string, string> = {
  shoutrrr: 'Shoutrrr (Telegram, Slack, etc.)',
  greenapi: 'GreenAPI (WhatsApp)',
  waweb: 'WaWeb (WhatsApp)',
}

type FormState = { name: string; type: string; config: Record<string, string> }

function parseConfig(raw: string): Record<string, string> {
  try {
    const parsed = JSON.parse(raw)
    // Ensure all values are strings
    return Object.fromEntries(Object.entries(parsed).map(([k, v]) => [k, String(v ?? '')]))
  } catch {
    return {}
  }
}

function ChannelForm({
  title,
  initial,
  isPending,
  error,
  onSave,
  onCancel,
  lockType,
}: {
  title: string
  initial: FormState
  isPending: boolean
  error?: string
  onSave: (f: FormState) => void
  onCancel: () => void
  lockType?: boolean
}) {
  const [form, setForm] = useState<FormState>(initial)

  const handleTypeChange = (type: string) => {
    // When changing type, keep keys that exist in the new default config
    const newDefaults = defaultConfigs[type] ?? {}
    const merged = { ...newDefaults }
    setForm({ ...form, type, config: merged })
  }

  // Merge config keys: show all keys from the default schema for this type,
  // plus any existing keys from the saved config that aren't in the defaults.
  const configKeys = Array.from(new Set([
    ...Object.keys(defaultConfigs[form.type] ?? {}),
    ...Object.keys(form.config),
  ]))

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-base">{title}</CardTitle>
        <Button variant="ghost" size="icon" onClick={onCancel}><X className="h-4 w-4" /></Button>
      </CardHeader>
      <CardContent className="space-y-3">
        <div>
          <label className="text-xs text-muted-foreground block mb-1">Name</label>
          <Input
            placeholder="Channel name"
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            autoFocus
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground block mb-1">Type</label>
          <select
            className="w-full border border-input rounded-md px-3 py-2 text-sm bg-background disabled:opacity-50"
            value={form.type}
            disabled={lockType}
            onChange={(e) => handleTypeChange(e.target.value)}
          >
            {Object.entries(typeLabels).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
          </select>
          {lockType && <p className="text-xs text-muted-foreground mt-1">Type cannot be changed after creation.</p>}
        </div>
        {configKeys.map((k) => (
          <div key={k}>
            <label className="text-xs text-muted-foreground block mb-1">{k}</label>
            <Input
              value={form.config[k] ?? ''}
              type={k.toLowerCase().includes('token') || k === 'auth' || k === 'url' ? 'password' : 'text'}
              placeholder={defaultConfigs[form.type]?.[k] !== undefined ? `${k}…` : ''}
              onChange={(e) => setForm({ ...form, config: { ...form.config, [k]: e.target.value } })}
            />
          </div>
        ))}
        {error && <p className="text-destructive text-sm">{error}</p>}
        <div className="flex gap-2 pt-1">
          <Button onClick={() => onSave(form)} disabled={isPending || !form.name}>
            {isPending ? 'Saving…' : 'Save'}
          </Button>
          <Button variant="outline" onClick={onCancel}>Cancel</Button>
        </div>
      </CardContent>
    </Card>
  )
}

export default function Channels() {
  const qc = useQueryClient()
  const [showAdd, setShowAdd] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [testResult, setTestResult] = useState<Record<string, string>>({})

  const { data: channels, isLoading } = useQuery({ queryKey: ['channels'], queryFn: api.channels.list })

  const createMutation = useMutation({
    mutationFn: (f: FormState) =>
      api.channels.create({ name: f.name, type: f.type as Channel['type'], config: f.config }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['channels'] })
      setShowAdd(false)
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, f }: { id: string; f: FormState }) =>
      api.channels.update(id, { name: f.name, type: f.type as Channel['type'], config: f.config, enabled: true }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['channels'] })
      setEditingId(null)
    },
  })

  const deleteMutation = useMutation({
    mutationFn: api.channels.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['channels'] }),
  })

  const testMutation = useMutation({
    mutationFn: api.channels.test,
    onSuccess: (_, id) => setTestResult((r) => ({ ...r, [id]: '✓ sent' })),
    onError: (err, id) => setTestResult((r) => ({ ...r, [id]: `✗ ${(err as Error).message}` })),
  })

  if (isLoading) return <div className="text-muted-foreground py-8 text-center">Loading…</div>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Notification Channels</h1>
        <Button size="sm" onClick={() => { setShowAdd(true); setEditingId(null) }}>
          <Plus className="h-4 w-4" /> Add Channel
        </Button>
      </div>

      {showAdd && (
        <ChannelForm
          title="New Channel"
          initial={{ name: '', type: 'shoutrrr', config: defaultConfigs.shoutrrr }}
          isPending={createMutation.isPending}
          error={createMutation.isError ? (createMutation.error as Error).message : undefined}
          onSave={(f) => createMutation.mutate(f)}
          onCancel={() => setShowAdd(false)}
        />
      )}

      {channels?.length === 0 && !showAdd && (
        <div className="text-center py-16 text-muted-foreground">
          <Bell className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p className="font-medium">No notification channels</p>
          <p className="text-sm mt-1">Add a channel to get alerted when new certificates appear.</p>
        </div>
      )}

      <div className="divide-y border rounded-lg overflow-hidden bg-card">
        {channels?.map((ch) => (
          <div key={ch.id}>
            {/* Edit form inline */}
            {editingId === ch.id ? (
              <div className="p-2">
                <ChannelForm
                  title={`Edit — ${ch.name}`}
                  initial={{ name: ch.name, type: ch.type, config: parseConfig(ch.config) }}
                  isPending={updateMutation.isPending}
                  error={updateMutation.isError ? (updateMutation.error as Error).message : undefined}
                  onSave={(f) => updateMutation.mutate({ id: ch.id, f })}
                  onCancel={() => setEditingId(null)}
                  lockType
                />
              </div>
            ) : (
              <div className="flex items-center gap-3 p-4">
                <div className="flex-1 min-w-0">
                  <div className="font-medium">{ch.name}</div>
                  <Badge variant="outline" className="mt-1">{ch.type}</Badge>
                </div>
                {testResult[ch.id] && (
                  <span className="text-xs text-muted-foreground shrink-0">{testResult[ch.id]}</span>
                )}
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => testMutation.mutate(ch.id)}
                  disabled={testMutation.isPending}
                  title="Send test notification"
                >
                  <Send className="h-3 w-3 mr-1" /> Test
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => { setEditingId(ch.id); setShowAdd(false) }}
                  title="Edit channel"
                >
                  <Pencil className="h-3 w-3" />
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => { if (confirm(`Delete "${ch.name}"?`)) deleteMutation.mutate(ch.id) }}
                  title="Delete channel"
                >
                  <Trash2 className="h-3 w-3" />
                </Button>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
