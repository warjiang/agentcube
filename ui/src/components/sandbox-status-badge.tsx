import { Badge } from '@/components/ui/badge'
import type { SandboxStatus } from '@/lib/types'

interface SandboxStatusBadgeProps {
  status: SandboxStatus | string
}

const statusConfig: Record<string, { bg: string; text: string; label: string }> = {
  running: { bg: '#ebf5ff', text: '#0068d6', label: 'Running' },
  creating: { bg: '#ebf5ff', text: '#666666', label: 'Creating' },
  unknown: { bg: '#ebebeb', text: '#666666', label: 'Unknown' },
}

export function SandboxStatusBadge({ status }: SandboxStatusBadgeProps) {
  const config = statusConfig[status] || statusConfig.unknown

  return (
    <Badge
      style={{
        backgroundColor: config.bg,
        color: config.text,
      }}
    >
      {config.label}
    </Badge>
  )
}
