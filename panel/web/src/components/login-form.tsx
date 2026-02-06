import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { zodResolver } from "@hookform/resolvers/zod"
import { useTranslation } from "react-i18next"
import { useForm } from "react-hook-form"
import { z } from "zod"

type LoginValues = {
  username: string
  password: string
}

export function LoginForm({
  className,
  isPending,
  errorMessage,
  onLogin,
  ...props
}: React.ComponentProps<"div"> & {
  isPending?: boolean
  errorMessage?: string | null
  onLogin: (values: LoginValues) => void
}) {
  const { t } = useTranslation()

  const loginSchema = z.object({
    username: z.string().min(1, t("auth.usernameRequired")),
    password: z.string().min(1, t("auth.passwordRequired")),
  })

  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: "", password: "" },
  })

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">{t("auth.adminLoginTitle")}</CardTitle>
          <CardDescription>{t("auth.adminLoginSubtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={form.handleSubmit((values) => onLogin(values))}
          >
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="username">{t("auth.username")}</FieldLabel>
                <Input
                  id="username"
                  autoComplete="username"
                  {...form.register("username")}
                />
                <FieldError errors={[form.formState.errors.username]} />
              </Field>
              <Field>
                <FieldLabel htmlFor="password">{t("auth.password")}</FieldLabel>
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  {...form.register("password")}
                />
                <FieldError errors={[form.formState.errors.password]} />
              </Field>
              <Field>
                {errorMessage ? <FieldError>{errorMessage}</FieldError> : null}
                <Button type="submit" disabled={isPending}>
                  {isPending ? t("auth.loggingIn") : t("auth.login")}
                </Button>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
      <FieldDescription className="px-6 text-center">
        {t("auth.tokenStoredHint")}
      </FieldDescription>
    </div>
  )
}
