import { useQuery } from '@tanstack/react-query'
import { Globe, FileText, Bell, Clock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { api } from '@/lib/api'
import { formatDistanceToNow } from 'date-fns'

function StatCard({ title, value, icon: Icon }: { title: string; value: number | string; icon: React.ElementType }) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
      </CardContent>
    </Card>
  )
}

export default function Dashboard() {
  const { data: fqdns } = useQuery({ queryKey: ['fqdns'], queryFn: api.fqdns.list })
  const { data: certs } = useQuery({ queryKey: ['certs-recent'], queryFn: () => api.certificates.list({ page_size: 10 }) })
  const { data: channels } = useQuery({ queryKey: ['channels'], queryFn: api.channels.list })
  const { data: schedules } = useQuery({ queryKey: ['schedules'], queryFn: api.schedules.list })

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
        <StatCard title="FQDNs" value={fqdns?.length ?? '—'} icon={Globe} />
        <StatCard title="Certificates" value={certs?.total ?? '—'} icon={FileText} />
        <StatCard title="Channels" value={channels?.length ?? '—'} icon={Bell} />
        <StatCard title="Schedules" value={schedules?.length ?? '—'} icon={Clock} />
      </div>
      <Card>
        <CardHeader><CardTitle>Recent Certificates</CardTitle></CardHeader>
        <CardContent>
          {(!certs?.items || certs.items.length === 0) && (
            <p className="text-muted-foreground text-sm py-4">No certificates discovered yet. Add an FQDN to start monitoring.</p>
          )}
          <div className="divide-y">
            {certs?.items?.map((cert) => (
              <div key={cert.id} className="py-3 flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-4">
                <div className="font-mono text-sm font-medium flex-1">{cert.subject_cn}</div>
                <div className="text-xs text-muted-foreground">{cert.issuer_ca}</div>
                <Badge variant={new Date(cert.not_after) < new Date() ? 'destructive' : 'secondary'} className="text-xs w-fit">
                  expires {formatDistanceToNow(new Date(cert.not_after), { addSuffix: true })}
                </Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
