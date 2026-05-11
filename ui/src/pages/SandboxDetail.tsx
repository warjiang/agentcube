import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { getSandbox } from '../api/sandbox'
import { ArrowLeft, Clock, Server, Activity, ExternalLink } from 'lucide-react'

function StatusBadge({ status }: { status: string }) {
  const isRunning = status.toLowerCase().includes('running')
  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium ${
      isRunning ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-700'
    }`}>
      {status}
    </span>
  )
}

export default function SandboxDetail() {
  const { id } = useParams<{ id: string }>()

  const { data: sandbox, isLoading, error } = useQuery({
    queryKey: ['sandbox', id],
    queryFn: () => getSandbox(id!),
    enabled: !!id,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600" />
      </div>
    )
  }

  if (error || !sandbox) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
        Failed to load sandbox. Please try again.
      </div>
    )
  }

  const endpoints = Object.entries(sandbox.Endpoints || {}).map(([name, url]) => ({
    name,
    url,
  }))

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/dashboard"
          className="flex items-center gap-2 text-slate-600 hover:text-slate-900"
        >
          <ArrowLeft className="w-5 h-5" />
          Back to Dashboard
        </Link>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-2xl font-bold text-slate-900">{sandbox.Name}</h2>
                <p className="text-slate-500">{sandbox.SandboxNamespace}</p>
                <div className="flex items-center gap-4 mt-3">
                  <StatusBadge status={sandbox.Status} />
                  <span className="text-sm text-slate-500 flex items-center gap-1">
                    <Activity className="w-4 h-4" />
                    {sandbox.Kind}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200">
              <h3 className="text-lg font-medium text-slate-900">Information</h3>
            </div>
            <div className="p-6 space-y-4">
              <InfoRow label="Sandbox ID" value={sandbox.SandboxID} />
              <InfoRow label="Session ID" value={sandbox.SessionID} />
              <InfoRow label="Namespace" value={sandbox.SandboxNamespace} />
              <InfoRow label="Name" value={sandbox.Name} />
              <InfoRow label="Kind" value={sandbox.Kind} />
              <InfoRow label="Image" value={sandbox.Image} />
              <InfoRow
                label="Created At"
                value={new Date(sandbox.CreatedAt).toLocaleString()}
              />
              <InfoRow
                label="Expires At"
                value={new Date(sandbox.ExpiresAt).toLocaleString()}
              />
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-200">
              <h3 className="text-lg font-medium text-slate-900">Endpoints</h3>
            </div>
            <div className="p-6">
              {endpoints.length === 0 ? (
                <p className="text-sm text-slate-500">No endpoints available</p>
              ) : (
                <div className="space-y-3">
                  {endpoints.map((endpoint) => (
                    <a
                      key={endpoint.name}
                      href={endpoint.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center justify-between p-3 bg-slate-50 rounded-lg hover:bg-slate-100 transition-colors"
                    >
                      <div>
                        <p className="text-sm font-medium text-slate-900">{endpoint.name}</p>
                        <p className="text-xs text-slate-500 truncate max-w-[200px]">
                          {endpoint.url}
                        </p>
                      </div>
                      <ExternalLink className="w-4 h-4 text-slate-400" />
                    </a>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between">
      <span className="text-sm text-slate-500">{label}</span>
      <span className="text-sm text-slate-900 text-right max-w-[250px] break-all">
        {value}
      </span>
    </div>
  )
}
