import { Link, Outlet, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { useAuthStore } from "@/store/auth"

const navItems = [
  { to: "/", label: "仪表盘" },
  { to: "/users", label: "用户" },
  { to: "/groups", label: "分组" },
  { to: "/nodes", label: "节点" },
  { to: "/inbounds", label: "入站" },
  { to: "/subscriptions", label: "订阅" },
  { to: "/settings", label: "设置" },
]

export function AppLayout() {
  const clearToken = useAuthStore((state) => state.clearToken)
  const navigate = useNavigate()

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-6">
          <span className="text-sm font-semibold tracking-wide text-slate-900">
            SBoard Panel
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              clearToken()
              navigate("/login", { replace: true })
            }}
          >
            退出登录
          </Button>
        </div>
      </header>
      <div className="mx-auto grid max-w-7xl grid-cols-1 gap-4 px-6 py-6 md:grid-cols-[220px_1fr]">
        <aside className="rounded-xl border border-slate-200 bg-white p-3">
          <nav className="space-y-1">
            {navItems.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className="block rounded-lg px-3 py-2 text-sm text-slate-700 transition-colors hover:bg-slate-100"
              >
                {item.label}
              </Link>
            ))}
          </nav>
        </aside>
        <main className="rounded-xl border border-slate-200 bg-white p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
