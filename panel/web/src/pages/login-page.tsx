import { useMutation, useQuery } from "@tanstack/react-query"
import { Navigate, useLocation, useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { BootstrapForm } from "@/components/bootstrap-form"
import { LoginForm } from "@/components/login-form"
import { bootstrapAdmin, getBootstrapStatus, loginAdmin } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { useAuthStore } from "@/store/auth"

export function LoginPage() {
  const { t } = useTranslation()
  const token = useAuthStore((state) => state.token)
  const setToken = useAuthStore((state) => state.setToken)
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? "/"

  const statusQuery = useQuery({
    queryKey: ["admin-bootstrap-status"],
    queryFn: getBootstrapStatus,
    retry: false,
  })

  const needsSetup = statusQuery.data?.needs_setup ?? false

  const loginMutation = useMutation({
    mutationFn: loginAdmin,
    onSuccess: (data) => {
      setToken(data.token)
      navigate(from, { replace: true })
    },
  })

  const bootstrapMutation = useMutation({
    mutationFn: bootstrapAdmin,
    onSuccess: async (_data, vars) => {
      // Auto-login after successful bootstrap.
      loginMutation.mutate({ username: vars.username, password: vars.password })
    },
  })

  if (token) {
    return <Navigate to={from} replace />
  }

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="w-full max-w-sm">
        {statusQuery.isLoading ? (
          <div className="text-muted-foreground text-sm">{t("auth.loading")}</div>
        ) : statusQuery.isError ? (
          <div className="text-destructive text-sm">
            {statusQuery.error instanceof ApiError
              ? statusQuery.error.message
              : t("auth.requestFailedHint")}
          </div>
        ) : needsSetup ? (
          <BootstrapForm
            isPending={bootstrapMutation.isPending || loginMutation.isPending}
            errorMessage={
              bootstrapMutation.isError
                ? bootstrapMutation.error instanceof ApiError
                  ? bootstrapMutation.error.message
                  : t("auth.bootstrapFailed")
                : loginMutation.isError
                  ? loginMutation.error instanceof ApiError
                    ? loginMutation.error.message
                    : t("auth.loginFailed")
                  : null
            }
            onBootstrap={(values) => bootstrapMutation.mutate(values)}
          />
        ) : (
          <LoginForm
            isPending={loginMutation.isPending}
            errorMessage={
              loginMutation.isError
                ? loginMutation.error instanceof ApiError
                  ? loginMutation.error.message
                  : t("auth.loginFailed")
                : null
            }
            onLogin={(values) => loginMutation.mutate(values)}
          />
        )}
      </div>
    </div>
  )
}
