import { Routes, Route, Link, useNavigate, useLocation } from 'react-router-dom'
import { LayoutDashboard, Server, Activity } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import SandboxDetail from './pages/SandboxDetail'
import { useState } from 'react'

function App() {
  return (
    <div className="flex h-screen bg-slate-100">
      <Sidebar />
      <div className="flex-1 overflow-auto">
        <Header />
        <main className="p-6">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/sandbox/:id" element={<SandboxDetail />} />
          </Routes>
        </main>
      </div>
    </div>
  )
}

function Sidebar() {
  const location = useLocation()
  
  const navItems = [
    { name: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
  ]

  return (
    <aside className="w-64 bg-slate-900 text-white flex flex-col">
      <div className="p-6 border-b border-slate-800">
        <div className="flex items-center gap-2">
          <Server className="w-8 h-8 text-indigo-400" />
          <span className="text-xl font-bold">AgentCube</span>
        </div>
        <p className="text-slate-400 text-sm mt-1">Sandbox Management</p>
      </div>
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => (
          <Link
            key={item.name}
            to={item.path}
            className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
              location.pathname.startsWith(item.path)
                ? 'bg-indigo-600 text-white'
                : 'text-slate-400 hover:bg-slate-800 hover:text-white'
            }`}
          >
            <item.icon className="w-5 h-5" />
            {item.name}
          </Link>
        ))}
      </nav>
    </aside>
  )
}

function Header() {
  const navigate = useNavigate()
  const [search, setSearch] = useState('')

  return (
    <header className="bg-white border-b border-slate-200 px-6 py-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Sandbox Management</h1>
          <p className="text-slate-500">Monitor and manage your sandboxes</p>
        </div>
        <div className="flex items-center gap-4">
          <div className="relative">
            <input
              type="text"
              placeholder="Search sandboxes..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
            />
            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">
              <Activity className="w-4 h-4" />
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}

export default App
