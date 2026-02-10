import type { ReactNode } from "react"

import { cn } from "@/lib/utils"

type PageHeaderProps = {
  title: string
  description?: string
  action?: ReactNode
  className?: string
}

export function PageHeader({
  title,
  description,
  action,
  className,
}: PageHeaderProps) {
  return (
    <header
      className={cn(
        "rounded-xl border border-border/80 bg-card px-4 py-4 shadow-[0_1px_0_0_rgba(255,255,255,0.35)_inset,0_14px_34px_-28px_rgba(0,0,0,0.45)] dark:shadow-[0_1px_0_0_rgba(255,255,255,0.08)_inset,0_20px_40px_-32px_rgba(0,0,0,0.85)] md:px-5 md:py-5",
        className,
      )}
    >
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
          {description ? <p className="mt-1 text-sm text-muted-foreground">{description}</p> : null}
        </div>
        {action ? <div className="shrink-0">{action}</div> : null}
      </div>
    </header>
  )
}
