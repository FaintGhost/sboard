import { describe, expect, it } from "vitest";

import { tableColumnLayout } from "./table-spacing";

describe("tableColumnLayout", () => {
  it("Given 七列表格且最后是操作列 When 使用节点列表布局 Then 前六列等分且操作列固定宽度", () => {
    const layout = tableColumnLayout.sevenActionIcon;
    expect(layout.tableClass).toBe("min-w-[920px] md:table-fixed");
    expect(layout.headFirst).toBe("md:w-[calc((100%-3rem)/6)]");
    expect(layout.headMiddle).toBe("md:w-[calc((100%-3rem)/6)]");
    expect(layout.headLast).toBe("w-12");
  });

  it("Given 三列表格无操作列 When 使用等距布局 Then 三列等宽", () => {
    const layout = tableColumnLayout.threeEven;
    expect(layout.tableClass).toBe("min-w-[520px] md:table-fixed");
    expect(layout.headFirst).toBe("md:w-1/3");
    expect(layout.headMiddle).toBe("md:w-1/3");
    expect(layout.headLast).toBe("md:w-1/3");
  });
});
