import { cn } from "@/lib/utils";

type StatusDotProps = {
  status: string;
  labelOnline: string;
  labelOffline: string;
  labelUnknown: string;
  className?: string;
};

export function StatusDot(props: StatusDotProps) {
  const s = (props.status || "").toLowerCase();
  const isOnline = s === "online";
  const isOffline = s === "offline";

  const label = isOnline ? props.labelOnline : isOffline ? props.labelOffline : props.labelUnknown;

  const dot = isOnline ? "bg-emerald-500" : isOffline ? "bg-destructive" : "bg-muted-foreground";
  const glow = isOnline
    ? "bg-emerald-400/70"
    : isOffline
      ? "bg-destructive/70"
      : "bg-muted-foreground/60";
  const ripple = isOnline
    ? "bg-emerald-400/25"
    : isOffline
      ? "bg-destructive/25"
      : "bg-muted-foreground/20";

  return (
    <span className={cn("inline-flex items-center gap-2", props.className)}>
      <span
        className="relative inline-flex size-3 shrink-0"
        data-status={isOnline ? "online" : isOffline ? "offline" : "unknown"}
        aria-label={label}
      >
        <span
          className={cn(
            "absolute inset-0 rounded-full motion-reduce:animate-none",
            ripple,
            "animate-[sboard-status-ripple_1.6s_ease-out_infinite]",
          )}
        />
        <span
          className={cn(
            "absolute inset-0 rounded-full blur-[5px] motion-reduce:animate-none",
            glow,
            "animate-[sboard-status-breathe_1.8s_ease-in-out_infinite]",
          )}
        />
        <span
          className={cn(
            "relative size-3 rounded-full",
            dot,
            "shadow-[0_0_0_1px_rgba(0,0,0,0.08),0_6px_14px_rgba(2,6,23,0.08)]",
          )}
        />
      </span>
      <span className="text-xs text-muted-foreground">{label}</span>
    </span>
  );
}
