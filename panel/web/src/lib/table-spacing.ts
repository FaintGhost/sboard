export type TableSpacingRule = {
  headFirst: string;
  headMiddle: string;
  headLast: string;
  cellFirst: string;
  cellMiddle: string;
  cellLast: string;
};

export type TableColumnLayoutRule = {
  tableClass: string;
  headFirst: string;
  headMiddle: string;
  headLast: string;
};

export const tableColumnSpacing: Record<
  "three" | "four" | "five" | "six" | "seven",
  TableSpacingRule
> = {
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
  seven: {
    headFirst: "pl-6",
    headMiddle: "px-3",
    headLast: "pr-6",
    cellFirst: "pl-6",
    cellMiddle: "px-3",
    cellLast: "pr-6",
  },
};

export const tableColumnLayout = {
  threeEven: {
    tableClass: "table-fixed",
    headFirst: "w-1/3",
    headMiddle: "w-1/3",
    headLast: "w-1/3",
  },
  fourActionIcon: {
    tableClass: "table-fixed",
    headFirst: "w-[calc((100%-3rem)/3)]",
    headMiddle: "w-[calc((100%-3rem)/3)]",
    headLast: "w-12",
  },
  fourActionWide: {
    tableClass: "table-fixed",
    headFirst: "w-[calc((100%-8.75rem)/3)]",
    headMiddle: "w-[calc((100%-8.75rem)/3)]",
    headLast: "w-[140px]",
  },
  fiveActionIcon: {
    tableClass: "table-fixed",
    headFirst: "w-[calc((100%-3rem)/4)]",
    headMiddle: "w-[calc((100%-3rem)/4)]",
    headLast: "w-12",
  },
  sixActionIcon: {
    tableClass: "table-fixed",
    headFirst: "w-[calc((100%-3rem)/5)]",
    headMiddle: "w-[calc((100%-3rem)/5)]",
    headLast: "w-12",
  },
  sevenActionIcon: {
    tableClass: "table-fixed",
    headFirst: "w-[calc((100%-3rem)/6)]",
    headMiddle: "w-[calc((100%-3rem)/6)]",
    headLast: "w-12",
  },
  sevenActionButton: {
    tableClass: "table-fixed",
    headFirst: "w-[calc((100%-7rem)/6)]",
    headMiddle: "w-[calc((100%-7rem)/6)]",
    headLast: "w-28",
  },
} satisfies Record<string, TableColumnLayoutRule>;
