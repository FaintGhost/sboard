import * as React from "react"
import {
  IconInnerShadowTop,
  IconSettings,
  IconUsers,
  IconServer2,
  IconArrowsExchange,
  IconListCheck,
} from "@tabler/icons-react"
import { Link, useLocation } from "react-router-dom"

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

const navItems = [
  { to: "/", label: "仪表盘", icon: IconInnerShadowTop },
  { to: "/users", label: "用户", icon: IconUsers },
  { to: "/groups", label: "分组", icon: IconListCheck },
  { to: "/nodes", label: "节点", icon: IconServer2 },
  { to: "/inbounds", label: "入站", icon: IconArrowsExchange },
  { to: "/subscriptions", label: "订阅", icon: IconArrowsExchange },
  { to: "/settings", label: "设置", icon: IconSettings },
] as const

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
        <SidebarMenu className="px-2">
          {navItems.map((item) => {
            const active =
              item.to === "/"
                ? location.pathname === "/"
                : location.pathname.startsWith(item.to)
            const Icon = item.icon
            return (
              <SidebarMenuItem key={item.to}>
                <SidebarMenuButton asChild isActive={active} tooltip={item.label}>
                  <Link to={item.to}>
                    <Icon />
                    <span>{item.label}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            )
          })}
        </SidebarMenu>
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={{ name: "admin" }} onLogout={onLogout} />
      </SidebarFooter>
    </Sidebar>
  )
}
