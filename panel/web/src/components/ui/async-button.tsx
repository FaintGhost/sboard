import { Loader2 } from "lucide-react";
import { useEffect, useRef, useState, type ComponentProps, type ReactNode } from "react";

import { Button } from "@/components/ui/button";

type AsyncButtonProps = ComponentProps<typeof Button> & {
  pending?: boolean;
  pendingText?: ReactNode;
  pendingDelayMs?: number;
  pendingMinMs?: number;
  showSpinner?: boolean;
};

function usePendingIndicator(
  pending: boolean,
  pendingDelayMs: number,
  pendingMinMs: number,
): boolean {
  const [visible, setVisible] = useState(false);
  const shownAtRef = useRef(0);
  const showTimerRef = useRef<number | null>(null);
  const hideTimerRef = useRef<number | null>(null);

  useEffect(() => {
    if (hideTimerRef.current != null) {
      window.clearTimeout(hideTimerRef.current);
      hideTimerRef.current = null;
    }

    if (pending) {
      if (visible) return;
      showTimerRef.current = window.setTimeout(
        () => {
          showTimerRef.current = null;
          shownAtRef.current = Date.now();
          setVisible(true);
        },
        Math.max(0, pendingDelayMs),
      );
      return;
    }

    if (showTimerRef.current != null) {
      window.clearTimeout(showTimerRef.current);
      showTimerRef.current = null;
    }

    if (!visible) return;

    const elapsed = Date.now() - shownAtRef.current;
    const remain = Math.max(0, pendingMinMs - elapsed);
    hideTimerRef.current = window.setTimeout(() => {
      setVisible(false);
      hideTimerRef.current = null;
    }, remain);
  }, [pending, pendingDelayMs, pendingMinMs, visible]);

  useEffect(() => {
    return () => {
      if (showTimerRef.current != null) {
        window.clearTimeout(showTimerRef.current);
        showTimerRef.current = null;
      }
      if (hideTimerRef.current != null) {
        window.clearTimeout(hideTimerRef.current);
        hideTimerRef.current = null;
      }
    };
  }, []);

  return visible;
}

export function AsyncButton({
  pending = false,
  pendingText,
  pendingDelayMs = 140,
  pendingMinMs = 320,
  showSpinner = true,
  disabled,
  children,
  ...props
}: AsyncButtonProps) {
  const pendingVisible = usePendingIndicator(pending, pendingDelayMs, pendingMinMs);
  const isBusy = pending || pendingVisible;

  return (
    <Button {...props} disabled={Boolean(disabled) || isBusy} aria-busy={isBusy || undefined}>
      {pendingVisible && showSpinner ? (
        <Loader2 className="size-4 animate-spin" data-testid="async-button-spinner" />
      ) : null}
      {pendingVisible ? (pendingText ?? children) : children}
    </Button>
  );
}
