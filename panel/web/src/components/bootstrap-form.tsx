import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AsyncButton } from "@/components/ui/async-button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FieldHint } from "@/components/ui/field-hint";

type BootstrapValues = {
  setup_token: string;
  username: string;
  password: string;
  confirm_password: string;
};

export function BootstrapForm({
  isPending,
  errorMessage,
  onBootstrap,
}: {
  isPending: boolean;
  errorMessage: string | null;
  onBootstrap: (values: BootstrapValues) => void;
}) {
  const { t } = useTranslation();

  const schema = z
    .object({
      setup_token: z.string().min(1, t("auth.setupTokenRequired")),
      username: z.string().min(1, t("auth.usernameRequired")),
      password: z.string().min(8, t("auth.passwordMinLength")),
      confirm_password: z.string().min(1, t("auth.confirmPasswordRequired")),
    })
    .refine((v) => v.password === v.confirm_password, {
      message: t("auth.passwordsDoNotMatch"),
      path: ["confirm_password"],
    });

  const form = useForm<BootstrapValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      setup_token: "",
      username: "",
      password: "",
      confirm_password: "",
    },
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-xl">{t("auth.bootstrapTitle")}</CardTitle>
        <CardDescription>{t("auth.bootstrapSubtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        {errorMessage ? (
          <Alert variant="destructive" className="mb-4">
            <AlertTitle>{t("auth.bootstrapFailed")}</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        ) : null}

        <form onSubmit={form.handleSubmit((values) => onBootstrap(values))} className="grid gap-4">
          <div className="grid gap-2">
            <div className="flex items-center gap-1">
              <Label htmlFor="setup_token">{t("auth.setupTokenLabel")}</Label>
              <FieldHint label={t("auth.setupTokenLabel")}>{t("auth.setupTokenHint")}</FieldHint>
            </div>
            <Input
              id="setup_token"
              placeholder={t("auth.setupTokenPlaceholder")}
              autoComplete="off"
              {...form.register("setup_token")}
            />
            {form.formState.errors.setup_token ? (
              <p className="text-destructive text-sm">
                {form.formState.errors.setup_token.message}
              </p>
            ) : null}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="username">{t("auth.username")}</Label>
            <Input id="username" autoComplete="username" {...form.register("username")} />
            {form.formState.errors.username ? (
              <p className="text-destructive text-sm">{form.formState.errors.username.message}</p>
            ) : null}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="password">{t("auth.password")}</Label>
            <Input
              id="password"
              type="password"
              autoComplete="new-password"
              {...form.register("password")}
            />
            {form.formState.errors.password ? (
              <p className="text-destructive text-sm">{form.formState.errors.password.message}</p>
            ) : null}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="confirm_password">{t("auth.confirmPassword")}</Label>
            <Input
              id="confirm_password"
              type="password"
              autoComplete="new-password"
              {...form.register("confirm_password")}
            />
            {form.formState.errors.confirm_password ? (
              <p className="text-destructive text-sm">
                {form.formState.errors.confirm_password.message}
              </p>
            ) : null}
          </div>

          <AsyncButton
            type="submit"
            className="w-full"
            disabled={isPending}
            pending={isPending}
            pendingText={t("auth.creatingAdmin")}
          >
            {t("auth.createAdmin")}
          </AsyncButton>
        </form>
      </CardContent>
    </Card>
  );
}
