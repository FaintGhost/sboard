export type TableSpacingRule = {
  headFirst: string
  headMiddle: string
  headLast: string
  cellFirst: string
  cellMiddle: string
  cellLast: string
}

export const tableColumnSpacing: Record<"three" | "four" | "five" | "six", TableSpacingRule> = {
  three: {
    headFirst: "pl-6",
    headMiddle: "px-4",
    headLast: "pr-6",
    cellFirst: "pl-6",
    cellMiddle: "px-4",
    cellLast: "pr-6",
  },
  four: {
    headFirst: "pl-6",
    headMiddle: "px-4",
    headLast: "pr-6",
    cellFirst: "pl-6",
    cellMiddle: "px-4",
    cellLast: "pr-6",
  },
  five: {
    headFirst: "pl-6",
    headMiddle: "px-3",
    headLast: "pr-6",
    cellFirst: "pl-6",
    cellMiddle: "px-3",
    cellLast: "pr-6",
  },
  six: {
    headFirst: "pl-6",
    headMiddle: "px-3",
    headLast: "pr-6",
    cellFirst: "pl-6",
    cellMiddle: "px-3",
    cellLast: "pr-6",
  },
}
