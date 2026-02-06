import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation } from "@tanstack/react-query"
import { useForm } from "react-hook-form"
import { Navigate, useLocation, useNavigate } from "react-router-dom"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { loginAdmin } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { useAuthStore } from "@/store/auth"

const loginSchema = z.object({
  username: z.string().min(1, "请输入用户名"),
  password: z.string().min(1, "请输入密码"),
})

type LoginFormValues = z.infer<typeof loginSchema>

export function LoginPage() {
  const token = useAuthStore((state) => state.token)
  const setToken = useAuthStore((state) => state.setToken)
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: string } | null)?.from ?? "/"

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  })

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
    <div className="grid min-h-screen place-items-center bg-slate-50 px-6">
      <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900">
          管理员登录
        </h1>
        <p className="mt-2 text-sm text-slate-500">登录后可访问管理面板</p>

        <form
          className="mt-6 space-y-4"
          onSubmit={form.handleSubmit((values) => mutation.mutate(values))}
        >
          <div className="space-y-1">
            <label className="text-sm text-slate-700" htmlFor="username">
              用户名
            </label>
            <input
              id="username"
              className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
              autoComplete="username"
              {...form.register("username")}
            />
            {form.formState.errors.username?.message ? (
              <p className="text-xs text-red-600">
                {form.formState.errors.username.message}
              </p>
            ) : null}
          </div>

          <div className="space-y-1">
            <label className="text-sm text-slate-700" htmlFor="password">
              密码
            </label>
            <input
              id="password"
              type="password"
              className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none focus:border-slate-500"
              autoComplete="current-password"
              {...form.register("password")}
            />
            {form.formState.errors.password?.message ? (
              <p className="text-xs text-red-600">
                {form.formState.errors.password.message}
              </p>
            ) : null}
          </div>

          {mutation.isError ? (
            <p className="text-sm text-red-600">
              {mutation.error instanceof ApiError
                ? mutation.error.message
                : "登录失败，请稍后再试"}
            </p>
          ) : null}

          <Button className="w-full" type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? "登录中..." : "登录"}
          </Button>
        </form>
      </div>
    </div>
  )
}
