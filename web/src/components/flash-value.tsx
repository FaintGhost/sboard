import * as React from "react";

import { cn } from "@/lib/utils";

type FlashValueProps = {
  value: string;
  className?: string;
};

// FlashValue re-triggers a small "fade in" transition whenever `value` changes.
// This avoids hard UI jumps without pulling in heavy animation libs.
export function FlashValue(props: FlashValueProps) {
  const [flash, setFlash] = React.useState(false);

  React.useEffect(() => {
    setFlash(false);
    const raf = requestAnimationFrame(() => setFlash(true));
    const t = window.setTimeout(() => setFlash(false), 320);
    return () => {
      cancelAnimationFrame(raf);
      window.clearTimeout(t);
    };
  }, [props.value]);

  return (
    <span
      className={cn(
        "inline-block tabular-nums transition-opacity duration-300",
        flash ? "animate-in fade-in-0" : "opacity-100",
        props.className,
      )}
    >
      {props.value}
    </span>
  );
}
