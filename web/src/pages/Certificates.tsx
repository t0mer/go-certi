import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { FileText, ChevronLeft, ChevronRight } from 'lucide-react'
import { api } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { formatDistanceToNow, parseISO } from 'date-fns'

export default function Certificates() {
  const [page, setPage] = useState(1)
  const [fqdnFilter, setFqdnFilter] = useState('')
  const [search, setSearch] = useState('')
  const pageSize = 25

  const { data, isLoading } = useQuery({
    queryKey: ['certificates', page, fqdnFilter],
    queryFn: () => api.certificates.list({ fqdn: fqdnFilter || undefined, page, page_size: pageSize }),
  })

  const totalPages = Math.ceil((data?.total ?? 0) / pageSize)

  const filtered = !search
    ? data?.items
    : data?.items?.filter((c) =>
        c.subject_cn.toLowerCase().includes(search.toLowerCase()) ||
        c.issuer_ca.toLowerCase().includes(search.toLowerCase()) ||
        c.sans.some((s) => s.toLowerCase().includes(search.toLowerCase()))
      )

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Certificates</h1>

      <div className="flex flex-col sm:flex-row gap-2">
        <Input
          placeholder="Filter by FQDN..."
          value={fqdnFilter}
          onChange={(e) => { setFqdnFilter(e.target.value); setPage(1) }}
          className="sm:w-52"
        />
        <Input
          placeholder="Search CN / SAN / CA..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1"
        />
      </div>

      {isLoading && <div className="text-muted-foreground text-center py-8">Loading...</div>}

      {!isLoading && (!filtered || filtered.length === 0) && (
        <div className="text-center py-16 text-muted-foreground">
          <FileText className="h-12 w-12 mx-auto mb-4 opacity-30" />
          <p className="font-medium">No certificates found</p>
          <p className="text-sm mt-1">Add FQDNs and trigger a scan to discover certificates.</p>
        </div>
      )}

      <div className="divide-y border rounded-lg overflow-hidden bg-card">
        {filtered?.map((cert) => {
          const expiry = parseISO(cert.not_after)
          const expired = expiry < new Date()
          const expiringSoon = !expired && expiry < new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
          return (
            <div key={cert.id} className="p-4">
              <div className="flex flex-col sm:flex-row sm:items-start gap-2">
                <div className="flex-1 min-w-0">
                  <div className="font-mono font-medium truncate">{cert.subject_cn}</div>
                  <div className="text-sm text-muted-foreground truncate">{cert.issuer_ca}</div>
                  {cert.sans.length > 1 && (
                    <div className="text-xs text-muted-foreground mt-1 truncate">
                      +{cert.sans.length - 1} SANs: {cert.sans.slice(1, 4).join(', ')}{cert.sans.length > 4 ? '...' : ''}
                    </div>
                  )}
                </div>
                <div className="flex flex-col items-end gap-1 shrink-0">
                  <Badge variant={expired ? 'destructive' : expiringSoon ? 'secondary' : 'default'}>
                    {expired ? 'Expired' : `expires ${formatDistanceToNow(expiry, { addSuffix: true })}`}
                  </Badge>
                  <span className="text-xs text-muted-foreground">{cert.source}</span>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            Page {page} of {totalPages} ({data?.total} total)
          </span>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page === totalPages}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
