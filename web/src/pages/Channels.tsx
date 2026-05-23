import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Send, Bell } from 'lucide-react'
import { api } from '@/lib/api'
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

export default function Channels() {
  const qc = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [testResult, setTestResult] = useState<Record<string, string>>({})
  const [form, setForm] = useState({ name: '', type: 'shoutrrr', config: defaultConfigs.shoutrrr })

  const { data: channels, isLoading } = useQuery({ queryKey: ['channels'], queryFn: api.channels.list })

  const createMutation = useMutation({
    mutationFn: () => api.channels.create({ name: form.name, type: form.type as 'shoutrrr' | 'greenapi' | 'waweb', config: form.config }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['channels'] }); setShowForm(false); setForm({ name: '', type: 'shoutrrr', config: defaultConfigs.shoutrrr }) },
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

  if (isLoading) return <div className="text-muted-foreground py-8 text-center">Loading...</div>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Notification Channels</h1>
        <Button size="sm" onClick={() => setShowForm(!showForm)}>
          <Plus className="h-4 w-4" /> Add Channel
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardHeader><CardTitle className="text-base">New Channel</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <Input placeholder="Channel name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} autoFocus />
            <select
              className="w-full border border-input rounded-md px-3 py-2 text-sm bg-background"
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value, config: defaultConfigs[e.target.value] })}
            >
              {Object.entries(typeLabels).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select>
            {Object.entries(form.config).map(([k]) => (
              <div key={k}>
                <label className="text-xs text-muted-foreground block mb-1">{k}</label>
                <Input
                  value={form.config[k]}
                  type={k.toLowerCase().includes('token') || k === 'auth' ? 'password' : 'text'}
                  onChange={(e) => setForm({ ...form, config: { ...form.config, [k]: e.target.value } })}
                />
              </div>
            ))}
            {createMutation.isError && <p className="text-destructive text-sm">{(createMutation.error as Error).message}</p>}
            <div className="flex gap-2">
              <Button onClick={() => createMutation.mutate()} disabled={createMutation.isPending || !form.name}>
                {createMutation.isPending ? 'Creating...' : 'Create'}
              </Button>
              <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
            </div>
          </CardContent>
        </Card>
      )}

      {channels?.length === 0 && (
        <div className="text-center py-16 text-muted-foreground">
          <Bell className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p className="font-medium">No notification channels</p>
          <p className="text-sm mt-1">Add a channel to get alerted when new certificates appear.</p>
        </div>
      )}

      <div className="divide-y border rounded-lg overflow-hidden bg-card">
        {channels?.map((ch) => (
          <div key={ch.id} className="flex items-center gap-3 p-4">
            <div className="flex-1 min-w-0">
              <div className="font-medium">{ch.name}</div>
              <Badge variant="outline" className="mt-1">{ch.type}</Badge>
            </div>
            {testResult[ch.id] && <span className="text-xs text-muted-foreground">{testResult[ch.id]}</span>}
            <Button size="sm" variant="outline" onClick={() => testMutation.mutate(ch.id)} disabled={testMutation.isPending} title="Send test notification">
              <Send className="h-3 w-3 mr-1" /> Test
            </Button>
            <Button size="sm" variant="destructive" onClick={() => { if (confirm(`Delete "${ch.name}"?`)) deleteMutation.mutate(ch.id) }}>
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        ))}
      </div>
    </div>
  )
}
