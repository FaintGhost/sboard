import { renderHook, waitFor } from "@testing-library/react"
import { describe, expect, it } from "vitest"

import { useTableQueryTransition } from "./table-query-transition"

type HookProps = {
  filterKey: string
  rows: Array<{ id: number }> | undefined
  isLoading: boolean
  isFetching: boolean
  isError: boolean
}

describe("useTableQueryTransition", () => {
  it("keeps no-data stable when switching from empty filter to empty filter", async () => {
    const { result, rerender } = renderHook(
      (props: HookProps) => useTableQueryTransition(props),
      {
        initialProps: {
          filterKey: "all",
          rows: undefined,
          isLoading: true,
          isFetching: true,
          isError: false,
        } as HookProps,
      },
    )

    expect(result.current.showSkeleton).toBe(false)
    expect(result.current.showLoadingHint).toBe(true)

    rerender({
      filterKey: "manual_node_sync",
      rows: [],
      isLoading: false,
      isFetching: false,
      isError: false,
    })

    await waitFor(() => {
      expect(result.current.showNoData).toBe(true)
      expect(result.current.showSkeleton).toBe(false)
      expect(result.current.showLoadingHint).toBe(false)
    })

    rerender({
      filterKey: "manual_retry",
      rows: undefined,
      isLoading: true,
      isFetching: true,
      isError: false,
    })

    await waitFor(() => {
      expect(result.current.showNoData).toBe(true)
      expect(result.current.showSkeleton).toBe(false)
      expect(result.current.showLoadingHint).toBe(false)
    })
  })

  it("hides stale rows without skeleton when switching filters", async () => {
    const { result, rerender } = renderHook(
      (props: HookProps) => useTableQueryTransition(props),
      {
        initialProps: {
          filterKey: "all",
          rows: [{ id: 1 }, { id: 2 }],
          isLoading: false,
          isFetching: false,
          isError: false,
        } as HookProps,
      },
    )

    rerender({
      filterKey: "failed",
      rows: undefined,
      isLoading: true,
      isFetching: true,
      isError: false,
    })

    await waitFor(() => {
      expect(result.current.showSkeleton).toBe(false)
      expect(result.current.showLoadingHint).toBe(false)
      expect(result.current.visibleRows).toEqual([])
      expect(result.current.showNoData).toBe(false)
    })
  })
})
