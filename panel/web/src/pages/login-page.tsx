import { useMutation } from "@tanstack/react-query"
import { Navigate, useLocation, useNavigate } from "react-router-dom"

import { LoginForm } from "@/components/login-form"
import { loginAdmin } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { useAuthStore } from "@/store/auth"

export function LoginPage() {
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
                : "登录失败，请稍后再试"
              : null
          }
          onLogin={(values) => mutation.mutate(values)}
        />
      </div>
    </div>
  )
}
