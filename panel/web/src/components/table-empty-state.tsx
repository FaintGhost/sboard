import { Link } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { TableCell, TableRow } from "@/components/ui/table";

type TableEmptyStateProps = {
  colSpan: number;
  message: string;
  actionLabel?: string;
  actionTo?: string;
  className?: string;
};

export function TableEmptyState({
  colSpan,
  message,
  actionLabel,
  actionTo,
  className,
}: TableEmptyStateProps) {
  return (
    <TableRow>
      <TableCell colSpan={colSpan} className={className ?? "py-10 text-center"}>
        <div className="mx-auto flex max-w-sm flex-col items-center gap-2 px-4">
          <p className="text-sm text-muted-foreground">{message}</p>
          {actionLabel && actionTo ? (
            <Button asChild size="sm" variant="outline">
              <Link to={actionTo}>{actionLabel}</Link>
            </Button>
          ) : null}
        </div>
      </TableCell>
    </TableRow>
  );
}
