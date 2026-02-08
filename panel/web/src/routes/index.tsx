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
import { SettingsPage } from "@/pages/settings-page"
import { SyncJobsPage } from "@/pages/sync-jobs-page"
import { SubscriptionsPage } from "@/pages/subscriptions-page"
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
          { path: "/sync-jobs", element: <SyncJobsPage /> },
          { path: "/subscriptions", element: <SubscriptionsPage /> },
          { path: "/settings", element: <SettingsPage /> },
        ],
      },
    ],
  },
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
