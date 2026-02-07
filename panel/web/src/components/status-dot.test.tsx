import { render, screen } from "@testing-library/react"
import { describe, expect, it } from "vitest"

import { StatusDot } from "./status-dot"

describe("StatusDot", () => {
  it("renders online", () => {
    render(
      <StatusDot
        status="online"
        labelOnline="Online"
        labelOffline="Offline"
        labelUnknown="Unknown"
      />,
    )
    expect(screen.getByText("Online")).toBeInTheDocument()
    expect(screen.getByLabelText("Online")).toHaveAttribute("data-status", "online")
  })

  it("renders offline", () => {
    render(
      <StatusDot
        status="offline"
        labelOnline="Online"
        labelOffline="Offline"
        labelUnknown="Unknown"
      />,
    )
    expect(screen.getByText("Offline")).toBeInTheDocument()
    expect(screen.getByLabelText("Offline")).toHaveAttribute("data-status", "offline")
  })
})

