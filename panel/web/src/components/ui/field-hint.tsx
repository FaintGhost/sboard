import { type ReactNode } from "react";
import { Info } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

type FieldHintProps = {
  label: string;
  children: ReactNode;
  className?: string;
  contentClassName?: string;
};

export function FieldHint({ label, children, className, contentClassName }: FieldHintProps) {
  return (
    <TooltipProvider delayDuration={120}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            className={cn(
              "rounded-full text-slate-500 hover:bg-slate-100 hover:text-slate-700",
              className,
            )}
            aria-label={`${label}说明`}
          >
            <Info className="size-3.5" />
          </Button>
        </TooltipTrigger>
        <TooltipContent
          side="top"
          align="start"
          sideOffset={6}
          className="max-w-xs text-xs leading-5 text-balance"
        >
          <div className={cn("whitespace-normal", contentClassName)}>{children}</div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
