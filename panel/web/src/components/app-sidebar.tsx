import * as React from "react"
import { useTranslation } from "react-i18next"
import {
  IconInnerShadowTop,
  IconSettings,
  IconUsers,
  IconServer2,
  IconArrowsExchange,
  IconListCheck,
  IconCloud,
  IconRefresh,
} from "@tabler/icons-react"
import { Link, useLocation } from "react-router-dom"

import { NavDocuments } from "@/components/nav-documents"
import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function AppSidebar({
  onLogout,
  ...props
}: React.ComponentProps<typeof Sidebar> & { onLogout: () => void }) {
  const { t } = useTranslation()
  const location = useLocation()

  const data = {
    user: {
      name: "admin",
      email: "admin",
      avatar: "/avatars/shadcn.jpg",
    },
    navMain: [
      {
        title: t("nav.dashboard"),
        url: "/",
        icon: IconInnerShadowTop,
      },
      {
        title: t("nav.users"),
        url: "/users",
        icon: IconUsers,
      },
      {
        title: t("nav.groups"),
        url: "/groups",
        icon: IconListCheck,
      },
      {
        title: t("nav.nodes"),
        url: "/nodes",
        icon: IconServer2,
      },
      {
        title: t("nav.inbounds"),
        url: "/inbounds",
        icon: IconArrowsExchange,
      },
      {
        title: t("nav.syncJobs"),
        url: "/sync-jobs",
        icon: IconRefresh,
      },
    ],
    navSecondary: [
      {
        title: t("nav.settings"),
        url: "/settings",
        icon: IconSettings,
      },
    ],
    documents: [
      {
        name: t("nav.subscriptions"),
        url: "/subscriptions",
        icon: IconCloud,
      },
    ],
  }

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
              <SidebarMenuButton
                asChild
                className="data-[slot=sidebar-menu-button]:!p-1.5"
              >
              <Link to="/">
                <IconInnerShadowTop className="!size-5" />
                <span className="text-base font-semibold">SBoard</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain
          items={data.navMain.map((item) => ({
            ...item,
            isActive:
              item.url === "/"
                ? location.pathname === "/"
                : location.pathname.startsWith(item.url),
          }))}
        />
        <NavDocuments items={data.documents} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} onLogout={onLogout} />
      </SidebarFooter>
    </Sidebar>
  )
}
