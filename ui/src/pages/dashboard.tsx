import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { Clock, Server, Box, Activity, Eye } from 'lucide-react'
import { listSandboxes } from '../api/sandbox'
import { SandboxInfo, Kind } from '../types/sandbox'
import { useState } from 'react'

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

function KindIcon({ kind }: { kind: Kind }) {
  if (kind === Kind.AgentRuntime) {
    return <Activity className="w-5 h-5 text-blue-500" />
  }
  return <Box className="w-5 h-5 text-purple-500" />
}

export default function Dashboard() {
  const [filters, setFilters] = useState({
    namespace: '',
    kind: '',
  })

  const { data, isLoading, error } = useQuery({
    queryKey: ['sandboxes', filters],
    queryFn: () => listSandboxes({
      namespace: filters.namespace || undefined,
      kind: filters.kind || undefined,
    }),
    refetchInterval: 5000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
        Failed to load sandboxes. Please try again.
      </div>
    )
  }

  const sandboxes = data?.items || []
  const total = data?.total || 0

  const stats = {
    total: total,
    running: sandboxes.filter(s => s.Status.toLowerCase().includes('running')).length,
    agentRuntimes: sandboxes.filter(s => s.Kind === Kind.AgentRuntime).length,
    codeInterpreters: sandboxes.filter(s => s.Kind === Kind.CodeInterpreter).length,
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <StatCard
          title="Total Sandboxes"
          value={stats.total}
          icon={Server}
          color="bg-blue-500"
        />
        <StatCard
          title="Running"
          value={stats.running}
          icon={Activity}
          color="bg-green-500"
        />
        <StatCard
          title="Agent Runtimes"
          value={stats.agentRuntimes}
          icon={Activity}
          color="bg-blue-600"
        />
        <StatCard
          title="Code Interpreters"
          value={stats.codeInterpreters}
          icon={Box}
          color="bg-purple-600"
        />
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-4 flex flex-wrap gap-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Namespace</label>
          <input
            type="text"
            placeholder="Filter by namespace"
            value={filters.namespace}
            onChange={(e) => setFilters(f => ({ ...f, namespace: e.target.value }))}
            className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1">Kind</label>
          <select
            value={filters.kind}
            onChange={(e) => setFilters(f => ({ ...f, kind: e.target.value }))}
            className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
          >
            <option value="">All</option>
            <option value={Kind.AgentRuntime}>AgentRuntime</option>
            <option value={Kind.CodeInterpreter}>CodeInterpreter</option>
          </select>
        </div>
        <div className="flex-end ml-auto flex items-end">
          <button
            onClick={() => setFilters({ namespace: '', kind: '' })}
            className="px-4 py-2 text-slate-600 hover:text-slate-800"
          >
            Reset
          </button>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-slate-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-slate-50 border-b border-slate-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Sandbox
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Kind
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Image
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Created
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-slate-200">
              {sandboxes.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-slate-500">
                    No sandboxes found
                  </td>
                </tr>
              ) : (
                sandboxes.map((sandbox) => (
                  <SandboxRow key={sandbox.SandboxID} sandbox={sandbox} />
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function StatCard({
  title,
  value,
  icon: Icon,
  color,
}: {
  title: string
  value: number
  icon: any
  color: string
}) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-slate-600">{title}</p>
          <p className="text-2xl font-bold text-slate-900 mt-1">{value}</p>
        </div>
        <div className={`p-3 rounded-lg ${color} text-white`}>
          <Icon className="w-6 h-6" />
        </div>
      </div>
    </div>
  )
}

function SandboxRow({ sandbox }: { sandbox: SandboxInfo }) {
  return (
    <tr className="hover:bg-slate-50">
      <td className="px-6 py-4 whitespace-nowrap">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <KindIcon kind={sandbox.Kind} />
          </div>
          <div className="ml-4">
            <div className="text-sm font-medium text-slate-900">{sandbox.Name}</div>
            <div className="text-sm text-slate-500">{sandbox.SandboxNamespace}</div>
            <div className="text-xs text-slate-400">ID: {sandbox.SandboxID}</div>
          </div>
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
        {sandbox.Kind}
      </td>
      <td className="px-6 py-4 whitespace-nowrap">
        <StatusBadge status={sandbox.Status} />
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
        {sandbox.Image}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
        <div className="flex items-center">
          <Clock className="w-4 h-4 mr-1" />
          {new Date(sandbox.CreatedAt).toLocaleString()}
        </div>
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
        <Link
          to={`/sandbox/${sandbox.SandboxID}`}
          className="inline-flex items-center gap-2 text-indigo-600 hover:text-indigo-900"
        >
          <Eye className="w-4 h-4" />
          View
        </Link>
      </td>
    </tr>
  )
}
