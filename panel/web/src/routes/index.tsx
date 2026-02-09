import { lazy, Suspense, type ReactNode } from "react"
import {
  createBrowserRouter,
  Navigate,
  Outlet,
  RouterProvider,
  useLocation,
} from "react-router-dom"

import { AppLayout } from "@/layouts/app-layout"
import { useAuthStore } from "@/store/auth"

const DashboardPage = lazy(() => import("@/pages/dashboard-page").then((mod) => ({ default: mod.DashboardPage })))
const GroupsPage = lazy(() => import("@/pages/groups-page").then((mod) => ({ default: mod.GroupsPage })))
const InboundsPage = lazy(() => import("@/pages/inbounds-page").then((mod) => ({ default: mod.InboundsPage })))
const LoginPage = lazy(() => import("@/pages/login-page").then((mod) => ({ default: mod.LoginPage })))
const NodesPage = lazy(() => import("@/pages/nodes-page").then((mod) => ({ default: mod.NodesPage })))
const SettingsPage = lazy(() => import("@/pages/settings-page").then((mod) => ({ default: mod.SettingsPage })))
const SyncJobsPage = lazy(() => import("@/pages/sync-jobs-page").then((mod) => ({ default: mod.SyncJobsPage })))
const SubscriptionsPage = lazy(() => import("@/pages/subscriptions-page").then((mod) => ({ default: mod.SubscriptionsPage })))
const UsersPage = lazy(() => import("@/pages/users-page").then((mod) => ({ default: mod.UsersPage })))

function withSuspense(element: ReactNode) {
  return (
    <Suspense fallback={<div className="px-4 py-6 text-sm text-muted-foreground">…</div>}>
      {element}
    </Suspense>
  )
}

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
    element: withSuspense(<LoginPage />),
  },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: "/", element: withSuspense(<DashboardPage />) },
          { path: "/users", element: withSuspense(<UsersPage />) },
          { path: "/groups", element: withSuspense(<GroupsPage />) },
          { path: "/nodes", element: withSuspense(<NodesPage />) },
          { path: "/inbounds", element: withSuspense(<InboundsPage />) },
          { path: "/sync-jobs", element: withSuspense(<SyncJobsPage />) },
          { path: "/subscriptions", element: withSuspense(<SubscriptionsPage />) },
          { path: "/settings", element: withSuspense(<SettingsPage />) },
        ],
      },
    ],
  },
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
