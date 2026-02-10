import { useTranslation } from "react-i18next";

import { AsyncButton } from "@/components/ui/async-button";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ApiError } from "@/lib/api/client";
import type { User } from "@/lib/api/types";

type DisableUserDialogProps = {
  user: User | null;
  onClose: () => void;
  onConfirm: (userId: number) => void;
  isPending: boolean;
  isError: boolean;
  error: Error | null;
};

export function DisableUserDialog({
  user,
  onClose,
  onConfirm,
  isPending,
  isError,
  error,
}: DisableUserDialogProps) {
  const { t } = useTranslation();

  return (
    <Dialog open={!!user} onOpenChange={(open) => (!open ? onClose() : null)}>
      <DialogContent aria-label={t("users.disableUser")}>
        <DialogHeader>
          <DialogTitle>{t("users.disableUser")}</DialogTitle>
          <DialogDescription>
            {t("users.disableConfirm", { username: user?.username })}
          </DialogDescription>
        </DialogHeader>

        {isError ? (
          <p className="text-sm text-destructive">
            {error instanceof ApiError ? error.message : t("users.disableFailed")}
          </p>
        ) : null}

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t("common.cancel")}
          </Button>
          <AsyncButton
            variant="destructive"
            onClick={() => {
              if (!user) return;
              onConfirm(user.id);
            }}
            disabled={isPending}
            pending={isPending}
            pendingText={t("common.disabling")}
          >
            {t("common.confirm")}
          </AsyncButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
