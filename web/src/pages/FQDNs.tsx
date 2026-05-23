import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, RefreshCw, Globe } from 'lucide-react'
import { api, type FQDN } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'

export default function FQDNs() {
  const qc = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [newFQDN, setNewFQDN] = useState('')

  const { data: fqdns, isLoading } = useQuery({ queryKey: ['fqdns'], queryFn: api.fqdns.list })

  const createMutation = useMutation({
    mutationFn: (fqdn: string) => api.fqdns.create({ fqdn }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['fqdns'] }); setNewFQDN(''); setShowForm(false) },
  })

  const deleteMutation = useMutation({
    mutationFn: api.fqdns.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fqdns'] }),
  })

  const scanMutation = useMutation({ mutationFn: api.fqdns.scan })

  const toggleMutation = useMutation({
    mutationFn: ({ id, f }: { id: string; f: FQDN }) => api.fqdns.update(id, { ...f, enabled: !f.enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fqdns'] }),
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
            <form className="flex gap-2" onSubmit={(e) => { e.preventDefault(); if (newFQDN) createMutation.mutate(newFQDN) }}>
              <Input placeholder="example.com" value={newFQDN} onChange={(e) => setNewFQDN(e.target.value)} className="flex-1" autoFocus />
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
          <div key={f.id} className="flex flex-col sm:flex-row sm:items-center gap-3 p-4">
            <div className="flex-1 min-w-0">
              <div className="font-mono font-medium truncate">{f.fqdn}</div>
              <div className="flex flex-wrap gap-1.5 mt-1">
                <Badge variant={f.enabled ? 'default' : 'secondary'}>{f.enabled ? 'enabled' : 'disabled'}</Badge>
                {f.include_subdomains && <Badge variant="outline">subdomains</Badge>}
                {f.notifications_enabled && <Badge variant="outline">notifications</Badge>}
              </div>
            </div>
            <div className="flex gap-2 shrink-0">
              <Button size="sm" variant="outline" onClick={() => toggleMutation.mutate({ id: f.id, f })} disabled={toggleMutation.isPending}>
                {f.enabled ? 'Disable' : 'Enable'}
              </Button>
              <Button size="sm" variant="outline" onClick={() => scanMutation.mutate(f.id)} disabled={scanMutation.isPending} title="Trigger scan">
                <RefreshCw className={`h-3 w-3 ${scanMutation.isPending ? 'animate-spin' : ''}`} />
              </Button>
              <Button size="sm" variant="destructive" onClick={() => { if (confirm(`Delete ${f.fqdn}?`)) deleteMutation.mutate(f.id) }}>
                <Trash2 className="h-3 w-3" />
              </Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
