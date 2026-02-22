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
    headFirst: "pl-4 sm:pl-6",
    headMiddle: "px-2 sm:px-4",
    headLast: "pr-4 sm:pr-6",
    cellFirst: "pl-4 sm:pl-6",
    cellMiddle: "px-2 sm:px-4",
    cellLast: "pr-4 sm:pr-6",
  },
  four: {
    headFirst: "pl-4 sm:pl-6",
    headMiddle: "px-2 sm:px-4",
    headLast: "pr-4 sm:pr-6",
    cellFirst: "pl-4 sm:pl-6",
    cellMiddle: "px-2 sm:px-4",
    cellLast: "pr-4 sm:pr-6",
  },
  five: {
    headFirst: "pl-4 sm:pl-6",
    headMiddle: "px-2 sm:px-3",
    headLast: "pr-4 sm:pr-6",
    cellFirst: "pl-4 sm:pl-6",
    cellMiddle: "px-2 sm:px-3",
    cellLast: "pr-4 sm:pr-6",
  },
  six: {
    headFirst: "pl-4 sm:pl-6",
    headMiddle: "px-2 sm:px-3",
    headLast: "pr-4 sm:pr-6",
    cellFirst: "pl-4 sm:pl-6",
    cellMiddle: "px-2 sm:px-3",
    cellLast: "pr-4 sm:pr-6",
  },
  seven: {
    headFirst: "pl-4 sm:pl-6",
    headMiddle: "px-2 sm:px-3",
    headLast: "pr-4 sm:pr-6",
    cellFirst: "pl-4 sm:pl-6",
    cellMiddle: "px-2 sm:px-3",
    cellLast: "pr-4 sm:pr-6",
  },
};

export const tableColumnLayout = {
  threeEven: {
    tableClass: "min-w-[520px] md:table-fixed",
    headFirst: "md:w-1/3",
    headMiddle: "md:w-1/3",
    headLast: "md:w-1/3",
  },
  fourActionIcon: {
    tableClass: "min-w-[600px] md:table-fixed",
    headFirst: "md:w-[calc((100%-3rem)/3)]",
    headMiddle: "md:w-[calc((100%-3rem)/3)]",
    headLast: "w-12",
  },
  fourActionWide: {
    tableClass: "min-w-[680px] md:table-fixed",
    headFirst: "md:w-[calc((100%-8.75rem)/3)]",
    headMiddle: "md:w-[calc((100%-8.75rem)/3)]",
    headLast: "w-[140px]",
  },
  fiveActionIcon: {
    tableClass: "min-w-[700px] md:table-fixed",
    headFirst: "md:w-[calc((100%-3rem)/4)]",
    headMiddle: "md:w-[calc((100%-3rem)/4)]",
    headLast: "w-12",
  },
  sixActionIcon: {
    tableClass: "min-w-[820px] md:table-fixed",
    headFirst: "md:w-[calc((100%-3rem)/5)]",
    headMiddle: "md:w-[calc((100%-3rem)/5)]",
    headLast: "w-12",
  },
  sevenActionIcon: {
    tableClass: "min-w-[920px] md:table-fixed",
    headFirst: "md:w-[calc((100%-3rem)/6)]",
    headMiddle: "md:w-[calc((100%-3rem)/6)]",
    headLast: "w-12",
  },
  sevenActionButton: {
    tableClass: "min-w-[980px] md:table-fixed",
    headFirst: "md:w-[calc((100%-7rem)/6)]",
    headMiddle: "md:w-[calc((100%-7rem)/6)]",
    headLast: "w-28",
  },
} satisfies Record<string, TableColumnLayoutRule>;
