import { useEffect, useMemo, useRef, useState } from "react"

type UseTableQueryTransitionOptions<Row> = {
  filterKey: string
  rows: Row[] | undefined
  isLoading: boolean
  isFetching: boolean
  isError: boolean
}

type UseTableQueryTransitionResult<Row> = {
  visibleRows: Row[]
  showSkeleton: boolean
  showLoadingHint: boolean
  showNoData: boolean
  isTransitioning: boolean
}

export function useTableQueryTransition<Row>(
  options: UseTableQueryTransitionOptions<Row>,
): UseTableQueryTransitionResult<Row> {
  const {
    filterKey,
    rows,
    isLoading,
    isFetching,
    isError,
  } = options

  const [isSwitching, setIsSwitching] = useState(false)
  const [lastSettledRowCount, setLastSettledRowCount] = useState<number | null>(null)
  const prevFilterKeyRef = useRef(filterKey)

  const justChangedFilter = prevFilterKeyRef.current !== filterKey
  const effectiveSwitching = isSwitching || justChangedFilter

  useEffect(() => {
    if (!justChangedFilter) return
    prevFilterKeyRef.current = filterKey
    setIsSwitching(true)
  }, [filterKey, justChangedFilter])

  useEffect(() => {
    if (!isFetching && rows) {
      setLastSettledRowCount(rows.length)
    }
  }, [isFetching, rows])

  useEffect(() => {
    if (!effectiveSwitching) return
    if (isError) {
      setIsSwitching(false)
      return
    }
    if (!isFetching && rows) {
      setIsSwitching(false)
    }
  }, [effectiveSwitching, isError, isFetching, rows])

  return useMemo(() => {
    const isTransitioning = effectiveSwitching && isFetching
    const visibleRows = isTransitioning ? [] : (rows ?? [])

    // 用户要求筛选过渡不展示骨架
    const showSkeleton = false

    const keepNoDataVisibleDuringSwitch = isTransitioning && lastSettledRowCount === 0
    const showNoData =
      keepNoDataVisibleDuringSwitch ||
      (!isLoading && !isTransitioning && rows != null && visibleRows.length === 0)

    const showLoadingHint = isLoading && !isTransitioning

    return {
      visibleRows,
      showSkeleton,
      showLoadingHint,
      showNoData,
      isTransitioning,
    }
  }, [
    effectiveSwitching,
    isFetching,
    isLoading,
    lastSettledRowCount,
    rows,
  ])
}
