import { useMutation } from "@tanstack/react-query"
import { Navigate, useLocation, useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"

import { LoginForm } from "@/components/login-form"
import { loginAdmin } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { useAuthStore } from "@/store/auth"

export function LoginPage() {
  const { t } = useTranslation()
  const token = useAuthStore((state) => state.token)
  const setToken = useAuthStore((state) => state.setToken)
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? "/"

  const mutation = useMutation({
    mutationFn: loginAdmin,
    onSuccess: (data) => {
      setToken(data.token)
      navigate(from, { replace: true })
    },
  })

  if (token) {
    return <Navigate to={from} replace />
  }

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="w-full max-w-sm">
        <LoginForm
          isPending={mutation.isPending}
          errorMessage={
            mutation.isError
              ? mutation.error instanceof ApiError
                ? mutation.error.message
                : t("auth.loginFailed")
              : null
          }
          onLogin={(values) => mutation.mutate(values)}
        />
      </div>
    </div>
  )
}
