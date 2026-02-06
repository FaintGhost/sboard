import {
  createBrowserRouter,
  Navigate,
  Outlet,
  RouterProvider,
  useLocation,
} from "react-router-dom"

import { AppLayout } from "@/layouts/app-layout"
import { DashboardPage } from "@/pages/dashboard-page"
import { GroupsPage } from "@/pages/groups-page"
import { InboundsPage } from "@/pages/inbounds-page"
import { LoginPage } from "@/pages/login-page"
import { NodesPage } from "@/pages/nodes-page"
import { UsersPage } from "@/pages/users-page"
import { useAuthStore } from "@/store/auth"

function RequireAuth() {
  const token = useAuthStore((state) => state.token)
  const location = useLocation()

  if (!token) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }

  return <Outlet />
}

function PlaceholderPage({ title }: { title: string }) {
  return (
    <div className="px-4 lg:px-6">
      <div className="space-y-2">
        <h1 className="text-xl font-semibold text-slate-900">{title}</h1>
        <p className="text-sm text-slate-500">该页面在下一阶段继续实现。</p>
      </div>
    </div>
  )
}

const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: "/", element: <DashboardPage /> },
          { path: "/users", element: <UsersPage /> },
          { path: "/groups", element: <GroupsPage /> },
          { path: "/nodes", element: <NodesPage /> },
          { path: "/inbounds", element: <InboundsPage /> },
          { path: "/subscriptions", element: <PlaceholderPage title="订阅管理" /> },
          { path: "/settings", element: <PlaceholderPage title="系统设置" /> },
        ],
      },
    ],
  },
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
