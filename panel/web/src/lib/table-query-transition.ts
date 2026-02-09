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
  const prevFilterKeyRef = useRef(filterKey)
  const prevRowsRef = useRef<Row[]>([])

  const justChangedFilter = prevFilterKeyRef.current !== filterKey
  const effectiveSwitching = isSwitching || justChangedFilter

  useEffect(() => {
    if (!justChangedFilter) return
    prevFilterKeyRef.current = filterKey
    setIsSwitching(true)
  }, [filterKey, justChangedFilter])

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

  // Keep previous rows during transition to avoid flash
  useEffect(() => {
    if (rows && !isFetching) {
      prevRowsRef.current = rows
    }
  }, [rows, isFetching])

  return useMemo(() => {
    // During filter switch, keep showing previous rows until new data arrives
    const isTransitioning = effectiveSwitching && isFetching
    const visibleRows = isTransitioning ? prevRowsRef.current : (rows ?? [])

    // Only show skeleton on initial load (not during filter transitions)
    const showSkeleton = isLoading && !effectiveSwitching && prevRowsRef.current.length === 0

    const showNoData = !showSkeleton && !isTransitioning && rows != null && rows.length === 0

    return {
      visibleRows,
      showSkeleton,
      showLoadingHint: isFetching,
      showNoData,
      isTransitioning,
    }
  }, [
    effectiveSwitching,
    isFetching,
    isLoading,
    rows,
  ])
}
