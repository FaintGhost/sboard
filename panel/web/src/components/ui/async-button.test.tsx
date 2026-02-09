import { act, render, screen } from "@testing-library/react"
import { describe, expect, it, vi } from "vitest"

import { AsyncButton } from "./async-button"

describe("AsyncButton", () => {
  it("does not flicker text for very short pending", () => {
    vi.useFakeTimers()

    const { rerender } = render(
      <AsyncButton pending={false} pendingText="保存中...">
        保存
      </AsyncButton>,
    )

    expect(screen.getByRole("button")).toHaveTextContent("保存")

    rerender(
      <AsyncButton pending pendingText="保存中...">
        保存
      </AsyncButton>,
    )

    act(() => {
      vi.advanceTimersByTime(80)
    })

    rerender(
      <AsyncButton pending={false} pendingText="保存中...">
        保存
      </AsyncButton>,
    )

    expect(screen.getByRole("button")).toHaveTextContent("保存")
    expect(screen.queryByText("保存中...")).not.toBeInTheDocument()

    vi.useRealTimers()
  })

  it("shows pending text after delay and keeps minimum visibility", () => {
    vi.useFakeTimers()

    const { rerender } = render(
      <AsyncButton pending={false} pendingText="重试中..." pendingDelayMs={120} pendingMinMs={300}>
        重试
      </AsyncButton>,
    )

    rerender(
      <AsyncButton pending pendingText="重试中..." pendingDelayMs={120} pendingMinMs={300}>
        重试
      </AsyncButton>,
    )

    act(() => {
      vi.advanceTimersByTime(130)
    })
    expect(screen.getByRole("button")).toHaveTextContent("重试中...")

    rerender(
      <AsyncButton pending={false} pendingText="重试中..." pendingDelayMs={120} pendingMinMs={300}>
        重试
      </AsyncButton>,
    )

    act(() => {
      vi.advanceTimersByTime(120)
    })
    expect(screen.getByRole("button")).toHaveTextContent("重试中...")

    act(() => {
      vi.advanceTimersByTime(220)
    })
    expect(screen.getByRole("button")).toHaveTextContent("重试")

    vi.useRealTimers()
  })
})
