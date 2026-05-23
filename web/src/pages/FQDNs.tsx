import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, RefreshCw, Globe, Settings, X } from 'lucide-react'
import { api, type FQDN, type Channel } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const EVENT_LABELS: Record<string, string> = {
  new_cert:      'New certificate issued',
  expiring_soon: 'Expiring soon',
  expired:       'Expired',
  revoked:       'Revoked',
  ca_changed:    'CA changed',
}
const ALL_EVENTS = Object.keys(EVENT_LABELS)

interface ConfigState {
  channelIds: string[]
  events: string[]
  expiryDays: number
}

function FQDNConfigPanel({
  fqdn,
  channels,
  onSave,
  onCancel,
  isPending,
}: {
  fqdn: FQDN
  channels: Channel[]
  onSave: (cfg: ConfigState) => void
  onCancel: () => void
  isPending: boolean
}) {
  const [cfg, setCfg] = useState<ConfigState>({
    channelIds: fqdn.channel_ids ?? [],
    events: fqdn.notification_events?.length ? fqdn.notification_events : ['new_cert'],
    expiryDays: fqdn.expiry_threshold_days ?? 10,
  })

  const toggleChannel = (id: string) =>
    setCfg((c) => ({
      ...c,
      channelIds: c.channelIds.includes(id)
        ? c.channelIds.filter((x) => x !== id)
        : [...c.channelIds, id],
    }))

  const toggleEvent = (ev: string) =>
    setCfg((c) => ({
      ...c,
      events: c.events.includes(ev)
        ? c.events.filter((x) => x !== ev)
        : [...c.events, ev],
    }))

  return (
    <Card className="mx-2 mb-2 border-primary/30">
      <CardHeader className="flex flex-row items-center justify-between pb-2 pt-3 px-4">
        <CardTitle className="text-sm font-medium">Configure — {fqdn.fqdn}</CardTitle>
        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={onCancel}>
          <X className="h-3 w-3" />
        </Button>
      </CardHeader>
      <CardContent className="px-4 pb-4 space-y-4">

        {/* Channel selection */}
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-2">NOTIFICATION CHANNELS</p>
          {channels.length === 0 ? (
            <p className="text-xs text-muted-foreground">No channels configured. Add channels first.</p>
          ) : (
            <div className="space-y-1.5">
              {channels.map((ch) => (
                <label key={ch.id} className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={cfg.channelIds.includes(ch.id)}
                    onChange={() => toggleChannel(ch.id)}
                    className="rounded"
                  />
                  <span className="text-sm">{ch.name}</span>
                  <Badge variant="outline" className="text-xs py-0">{ch.type}</Badge>
                </label>
              ))}
            </div>
          )}
        </div>

        {/* Event selection */}
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-2">NOTIFY ON</p>
          <div className="space-y-1.5">
            {ALL_EVENTS.map((ev) => (
              <div key={ev}>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={cfg.events.includes(ev)}
                    onChange={() => toggleEvent(ev)}
                    className="rounded"
                  />
                  <span className="text-sm">{EVENT_LABELS[ev]}</span>
                </label>
                {ev === 'expiring_soon' && cfg.events.includes('expiring_soon') && (
                  <div className="flex items-center gap-2 ml-6 mt-1.5">
                    <label className="text-xs text-muted-foreground whitespace-nowrap">
                      Notify when expires within
                    </label>
                    <Input
                      type="number"
                      min={1}
                      max={365}
                      value={cfg.expiryDays}
                      onChange={(e) => setCfg({ ...cfg, expiryDays: Math.max(1, Number(e.target.value)) })}
                      className="h-7 w-20 text-xs"
                    />
                    <span className="text-xs text-muted-foreground">days</span>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        <div className="flex gap-2 pt-1">
          <Button size="sm" onClick={() => onSave(cfg)} disabled={isPending}>
            {isPending ? 'Saving…' : 'Save'}
          </Button>
          <Button size="sm" variant="outline" onClick={onCancel}>Cancel</Button>
        </div>
      </CardContent>
    </Card>
  )
}

export default function FQDNs() {
  const qc = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [newFQDN, setNewFQDN] = useState('')
  const [configuringId, setConfiguringId] = useState<string | null>(null)

  const { data: fqdns, isLoading } = useQuery({ queryKey: ['fqdns'], queryFn: api.fqdns.list })
  const { data: channels = [] } = useQuery({ queryKey: ['channels'], queryFn: api.channels.list })

  const createMutation = useMutation({
    mutationFn: (fqdn: string) => api.fqdns.create({ fqdn, notification_events: ['new_cert'], expiry_threshold_days: 10 }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['fqdns'] }); setNewFQDN(''); setShowForm(false) },
  })

  const deleteMutation = useMutation({
    mutationFn: api.fqdns.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fqdns'] }),
  })

  const scanMutation = useMutation({ mutationFn: api.fqdns.scan })

  const toggleMutation = useMutation({
    mutationFn: ({ id, f }: { id: string; f: FQDN }) =>
      api.fqdns.update(id, { ...f, enabled: !f.enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fqdns'] }),
  })

  const saveCfgMutation = useMutation({
    mutationFn: ({ id, f, cfg }: { id: string; f: FQDN; cfg: ConfigState }) =>
      api.fqdns.update(id, {
        ...f,
        channel_ids: cfg.channelIds,
        notification_events: cfg.events,
        expiry_threshold_days: cfg.expiryDays,
      }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['fqdns'] }); setConfiguringId(null) },
  })

  if (isLoading) return <div className="text-muted-foreground py-8 text-center">Loading...</div>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">FQDNs</h1>
        <Button onClick={() => setShowForm(!showForm)} size="sm">
          <Plus className="h-4 w-4" /> Add FQDN
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardContent className="pt-4">
            <form
              className="flex gap-2"
              onSubmit={(e) => { e.preventDefault(); if (newFQDN) createMutation.mutate(newFQDN) }}
            >
              <Input
                placeholder="example.com"
                value={newFQDN}
                onChange={(e) => setNewFQDN(e.target.value)}
                className="flex-1"
                autoFocus
              />
              <Button type="submit" disabled={createMutation.isPending || !newFQDN}>Add</Button>
              <Button type="button" variant="outline" onClick={() => { setShowForm(false); setNewFQDN('') }}>Cancel</Button>
            </form>
            {createMutation.isError && (
              <p className="text-destructive text-sm mt-2">{(createMutation.error as Error).message}</p>
            )}
          </CardContent>
        </Card>
      )}

      {fqdns?.length === 0 && (
        <div className="text-center py-16 text-muted-foreground">
          <Globe className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p className="font-medium">No FQDNs yet</p>
          <p className="text-sm mt-1">Add a domain to start monitoring Certificate Transparency logs.</p>
        </div>
      )}

      <div className="divide-y border rounded-lg overflow-hidden bg-card">
        {fqdns?.map((f) => (
          <div key={f.id}>
            {/* Row */}
            <div className="flex flex-col sm:flex-row sm:items-center gap-3 p-4">
              <div className="flex-1 min-w-0">
                <div className="font-mono font-medium truncate">{f.fqdn}</div>
                <div className="flex flex-wrap gap-1.5 mt-1">
                  <Badge variant={f.enabled ? 'default' : 'secondary'}>
                    {f.enabled ? 'enabled' : 'disabled'}
                  </Badge>
                  {f.include_subdomains && <Badge variant="outline">subdomains</Badge>}
                  {f.channel_ids?.length > 0 && (
                    <Badge variant="outline">{f.channel_ids.length} channel{f.channel_ids.length !== 1 ? 's' : ''}</Badge>
                  )}
                  {f.notification_events?.length > 0 && f.notifications_enabled && (
                    <Badge variant="outline">{f.notification_events.length} event{f.notification_events.length !== 1 ? 's' : ''}</Badge>
                  )}
                </div>
              </div>
              <div className="flex gap-2 shrink-0">
                <Button size="sm" variant="outline" onClick={() => toggleMutation.mutate({ id: f.id, f })} disabled={toggleMutation.isPending}>
                  {f.enabled ? 'Disable' : 'Enable'}
                </Button>
                <Button
                  size="sm"
                  variant={configuringId === f.id ? 'default' : 'outline'}
                  onClick={() => setConfiguringId(configuringId === f.id ? null : f.id)}
                  title="Configure channels and events"
                >
                  <Settings className="h-3 w-3" />
                </Button>
                <Button size="sm" variant="outline" onClick={() => scanMutation.mutate(f.id)} disabled={scanMutation.isPending} title="Trigger scan">
                  <RefreshCw className={`h-3 w-3 ${scanMutation.isPending ? 'animate-spin' : ''}`} />
                </Button>
                <Button size="sm" variant="destructive" onClick={() => { if (confirm(`Delete ${f.fqdn}?`)) deleteMutation.mutate(f.id) }}>
                  <Trash2 className="h-3 w-3" />
                </Button>
              </div>
            </div>

            {/* Inline config panel */}
            {configuringId === f.id && (
              <FQDNConfigPanel
                fqdn={f}
                channels={channels}
                onSave={(cfg) => saveCfgMutation.mutate({ id: f.id, f, cfg })}
                onCancel={() => setConfiguringId(null)}
                isPending={saveCfgMutation.isPending}
              />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
