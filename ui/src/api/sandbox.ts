import { SandboxInfo, ListSandboxesResponse } from '../types/sandbox'

const API_BASE = 'http://localhost:8080/v1'

export async function listSandboxes(options: {
  namespace?: string
  kind?: string
  limit?: number
  offset?: number
} = {}): Promise<ListSandboxesResponse> {
  const params = new URLSearchParams()
  if (options.namespace) params.set('namespace', options.namespace)
  if (options.kind) params.set('kind', options.kind)
  if (options.limit) params.set('limit', options.limit.toString())
  if (options.offset) params.set('offset', options.offset.toString())

  const response = await fetch(`${API_BASE}/sandboxes?${params.toString()}`)
  if (!response.ok) {
    throw new Error('Failed to list sandboxes')
  }
  return response.json()
}

export async function getSandbox(id: string): Promise<SandboxInfo> {
  const response = await fetch(`${API_BASE}/sandboxes/${id}`)
  if (!response.ok) {
    throw new Error('Failed to get sandbox')
  }
  return response.json()
}
