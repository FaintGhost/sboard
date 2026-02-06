import type { CSSProperties } from "react"
import { Outlet, useLocation, useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { useAuthStore } from "@/store/auth"

function titleForPath(t: (key: string) => string, pathname: string): string {
  if (pathname === "/") return t("nav.dashboard")
  if (pathname.startsWith("/users")) return t("nav.users")
  if (pathname.startsWith("/groups")) return t("nav.groups")
  if (pathname.startsWith("/nodes")) return t("nav.nodes")
  if (pathname.startsWith("/inbounds")) return t("nav.inbounds")
  if (pathname.startsWith("/subscriptions")) return t("nav.subscriptions")
  if (pathname.startsWith("/settings")) return t("nav.settings")
  return t("app.title")
}

export function AppLayout() {
  const { t } = useTranslation()
  const clearToken = useAuthStore((state) => state.clearToken)
  const navigate = useNavigate()
  const location = useLocation()
  const title = titleForPath(t, location.pathname)

  return (
    <SidebarProvider
      className="bg-sidebar"
      style={{ "--header-height": "3.5rem" } as CSSProperties}
    >
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
