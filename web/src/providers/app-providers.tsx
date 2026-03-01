import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";
import { useState, type ReactNode } from "react";

import { ErrorBoundary } from "@/components/error-boundary";
import { rpcTransport } from "@/lib/rpc/transport";
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";

type AppProvidersProps = {
  children: ReactNode;
};

export function AppProviders({ children }: AppProvidersProps) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            retry: 1,
            refetchOnWindowFocus: false,
          },
        },
      }),
  );

  return (
    <ErrorBoundary>
      <TransportProvider transport={rpcTransport}>
        <QueryClientProvider client={queryClient}>
          <TooltipProvider delayDuration={120}>
            {children}
            <Toaster />
          </TooltipProvider>
        </QueryClientProvider>
      </TransportProvider>
    </ErrorBoundary>
  );
}
