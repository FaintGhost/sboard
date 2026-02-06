import * as React from "react"
import {
  IconInnerShadowTop,
  IconSettings,
  IconUsers,
  IconServer2,
  IconArrowsExchange,
  IconListCheck,
  IconCloud,
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

const data = {
  user: {
    name: "admin",
    email: "admin",
    avatar: "/avatars/shadcn.jpg",
  },
  navMain: [
    {
      title: "仪表盘",
      url: "/",
      icon: IconInnerShadowTop,
    },
    {
      title: "用户",
      url: "/users",
      icon: IconUsers,
    },
    {
      title: "分组",
      url: "/groups",
      icon: IconListCheck,
    },
    {
      title: "节点",
      url: "/nodes",
      icon: IconServer2,
    },
    {
      title: "入站",
      url: "/inbounds",
      icon: IconArrowsExchange,
    },
  ],
  navSecondary: [
    {
      title: "系统设置",
      url: "/settings",
      icon: IconSettings,
    },
  ],
  documents: [
    {
      name: "订阅",
      url: "/subscriptions",
      icon: IconCloud,
    },
  ],
}

export function AppSidebar({
  onLogout,
  ...props
}: React.ComponentProps<typeof Sidebar> & { onLogout: () => void }) {
  const location = useLocation()

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
