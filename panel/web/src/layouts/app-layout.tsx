import type { CSSProperties } from "react"
import { Outlet, useLocation, useNavigate } from "react-router-dom"

import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { useAuthStore } from "@/store/auth"

function titleForPath(pathname: string): string {
  if (pathname === "/") return "仪表盘"
  if (pathname.startsWith("/users")) return "用户管理"
  if (pathname.startsWith("/groups")) return "分组管理"
  if (pathname.startsWith("/nodes")) return "节点管理"
  if (pathname.startsWith("/inbounds")) return "入站管理"
  if (pathname.startsWith("/subscriptions")) return "订阅管理"
  if (pathname.startsWith("/settings")) return "系统设置"
  return "面板"
}

export function AppLayout() {
  const clearToken = useAuthStore((state) => state.clearToken)
  const navigate = useNavigate()
  const location = useLocation()
  const title = titleForPath(location.pathname)

  return (
    <SidebarProvider style={{ "--header-height": "3.5rem" } as CSSProperties}>
      <AppSidebar
        variant="inset"
        onLogout={() => {
          clearToken()
          navigate("/login", { replace: true })
        }}
      />
      <SidebarInset>
        <SiteHeader title={title} />
        <div className="@container/main flex flex-1 flex-col gap-4 py-4">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
