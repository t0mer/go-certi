import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Star, Clock } from 'lucide-react'
import { api, type Schedule } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'

export default function Schedules() {
  const qc = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', cron_expr: '@every 2h' })

  const { data: schedules, isLoading } = useQuery({ queryKey: ['schedules'], queryFn: api.schedules.list })

  const createMutation = useMutation({
    mutationFn: () => api.schedules.create({ name: form.name, cron_expr: form.cron_expr }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['schedules'] }); setShowForm(false); setForm({ name: '', cron_expr: '@every 2h' }) },
  })

  const deleteMutation = useMutation({
    mutationFn: api.schedules.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedules'] }),
  })

  const setDefaultMutation = useMutation({
    mutationFn: (s: Schedule) => api.schedules.update(s.id, { ...s, is_default: true }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedules'] }),
  })

  if (isLoading) return <div className="text-muted-foreground py-8 text-center">Loading...</div>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Schedules</h1>
        <Button size="sm" onClick={() => setShowForm(!showForm)}>
          <Plus className="h-4 w-4" /> Add Schedule
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardContent className="pt-4 space-y-3">
            <Input placeholder="Name (e.g. Hourly)" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} autoFocus />
            <Input placeholder="Cron expression (e.g. @every 1h)" value={form.cron_expr} onChange={(e) => setForm({ ...form, cron_expr: e.target.value })} />
            <p className="text-xs text-muted-foreground">Examples: @every 1h · @every 4h · 0 0 */6 * * (every 6h) · @daily</p>
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

      <div className="divide-y border rounded-lg overflow-hidden bg-card">
        {schedules?.map((s) => (
          <div key={s.id} className="flex items-center gap-3 p-4">
            <Clock className="h-4 w-4 text-muted-foreground shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="font-medium flex items-center gap-2">
                {s.name}
                {s.is_default && <Star className="h-3.5 w-3.5 fill-yellow-400 text-yellow-400" />}
              </div>
              <code className="text-xs text-muted-foreground">{s.cron_expr}</code>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              {s.is_default ? (
                <Badge>default</Badge>
              ) : (
                <>
                  <Button size="sm" variant="outline" onClick={() => setDefaultMutation.mutate(s)} disabled={setDefaultMutation.isPending}>
                    Set default
                  </Button>
                  <Button size="sm" variant="destructive" onClick={() => { if (confirm(`Delete "${s.name}"?`)) deleteMutation.mutate(s.id) }}>
                    <Trash2 className="h-3 w-3" />
                  </Button>
                </>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
