import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Download, Clock, AlertCircle, ExternalLink, X } from 'lucide-react'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'

const SKIP_KEY = 'go_certi_skipped_version'
const REMIND_KEY = 'go_certi_remind_at'
const ENABLED_KEY = 'go_certi_update_check'
const INTERVAL_KEY = 'go_certi_update_interval_h'

function shouldCheck(): boolean {
  if (localStorage.getItem(ENABLED_KEY) === 'false') return false
  const intervalH = Number(localStorage.getItem(INTERVAL_KEY) ?? '24')
  const lastKey = 'go_certi_last_check'
  const last = localStorage.getItem(lastKey)
  if (last) {
    const elapsed = (Date.now() - Number(last)) / 3600_000
    if (elapsed < intervalH) return false
  }
  localStorage.setItem(lastKey, String(Date.now()))
  return true
}

export function UpdateBanner() {
  const [dismissed, setDismissed] = useState(false)
  const [applying, setApplying] = useState(false)
  const [applyError, setApplyError] = useState('')
  const [applyStatus, setApplyStatus] = useState('')
  const [enabled, setEnabled] = useState(() => localStorage.getItem(ENABLED_KEY) !== 'false')

  // Expose a way for the Settings page to refresh this banner's enabled state
  useEffect(() => {
    const handler = () => setEnabled(localStorage.getItem(ENABLED_KEY) !== 'false')
    window.addEventListener('storage', handler)
    return () => window.removeEventListener('storage', handler)
  }, [])

  const { data: status } = useQuery({
    queryKey: ['update-status'],
    queryFn: api.updates.status,
    // Only fetch if update checks are enabled and interval has elapsed
    enabled: enabled && shouldCheck(),
    staleTime: Infinity,   // don't re-fetch on window focus
    retry: false,
  })

  if (!status?.update_available || dismissed) return null

  // Honor "skip this version"
  const skipped = localStorage.getItem(SKIP_KEY)
  if (skipped === status.latest_version) return null

  // Honor "remind later"
  const remindAt = localStorage.getItem(REMIND_KEY)
  if (remindAt && new Date(remindAt) > new Date()) return null

  const handleSkip = () => {
    localStorage.setItem(SKIP_KEY, status.latest_version)
    setDismissed(true)
  }

  const handleRemind = () => {
    const at = new Date(Date.now() + 24 * 3600_000).toISOString()
    localStorage.setItem(REMIND_KEY, at)
    setDismissed(true)
  }

  const handleApply = async () => {
    setApplying(true)
    setApplyError('')
    setApplyStatus('Downloading update…')
    try {
      await api.updates.apply()
      setApplyStatus('Update applied. Waiting for server to restart…')
      // Poll /healthz until the server comes back with the new binary
      let attempts = 0
      const poll = setInterval(async () => {
        attempts++
        try {
          const r = await fetch('/healthz')
          if (r.ok) {
            clearInterval(poll)
            setApplyStatus('Server restarted. Reloading…')
            setTimeout(() => window.location.reload(), 800)
          }
        } catch {}
        if (attempts > 30) {
          clearInterval(poll)
          setApplying(false)
          setApplyError('Server did not restart within 60 seconds — please reload manually.')
        }
      }, 2000)
    } catch (err) {
      setApplyError((err as Error).message)
      setApplying(false)
      setApplyStatus('')
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-card border rounded-xl shadow-xl max-w-lg w-full">
        {/* Header */}
        <div className="flex items-center justify-between px-5 pt-5 pb-3 border-b">
          <div className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-primary shrink-0" />
            <h2 className="font-semibold">Update available</h2>
          </div>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleRemind} title="Remind me later">
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Body */}
        <div className="px-5 py-4 space-y-3">
          <div className="flex gap-4 text-sm">
            <span>
              <span className="text-muted-foreground">Running: </span>
              <code className="bg-muted px-1.5 py-0.5 rounded text-xs">{status.current_version}</code>
            </span>
            <span>
              <span className="text-muted-foreground">Latest: </span>
              <code className="bg-muted px-1.5 py-0.5 rounded text-xs font-semibold">{status.latest_version}</code>
            </span>
          </div>

          {status.release_notes && (
            <details className="text-sm">
              <summary className="cursor-pointer text-muted-foreground hover:text-foreground">Release notes</summary>
              <pre className="mt-2 p-3 bg-muted rounded text-xs whitespace-pre-wrap max-h-40 overflow-auto">
                {status.release_notes}
              </pre>
            </details>
          )}

          {status.release_url && (
            <a href={status.release_url} target="_blank" rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs text-primary hover:underline">
              View on GitHub <ExternalLink className="h-3 w-3" />
            </a>
          )}

          {applyStatus && <p className="text-sm text-muted-foreground">{applyStatus}</p>}
          {applyError && <p className="text-sm text-destructive">{applyError}</p>}
        </div>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row gap-2 px-5 pb-5">
          <Button onClick={handleApply} disabled={applying} className="flex-1">
            <Download className="h-4 w-4" />
            {applying ? 'Updating…' : 'Update now'}
          </Button>
          <Button variant="outline" onClick={handleRemind} disabled={applying}>
            <Clock className="h-4 w-4" /> Remind in 24h
          </Button>
          <Button variant="ghost" onClick={handleSkip} disabled={applying} className="text-muted-foreground">
            Skip this version
          </Button>
        </div>
      </div>
    </div>
  )
}
