import { mount, flushPromises } from "@vue/test-utils";
import { describe, expect, it, vi } from "vitest";

const api = vi.hoisted(() => ({
  reconcileBatches: vi.fn(),
  reconcileDetails: vi.fn(),
  hedgeTasks: vi.fn(),
  riskEvents: vi.fn(),
  createManualHedge: vi.fn(),
  runReconcile: vi.fn(),
  cancelHedgeTask: vi.fn(),
  retryHedgeTask: vi.fn(),
  resolveRiskEvent: vi.fn(),
}));

vi.mock("@/api/liquidity", () => ({ liquidityApi: api }));
vi.mock("element-plus", () => ({
  ElMessage: { success: vi.fn(), warning: vi.fn() },
  ElMessageBox: { prompt: vi.fn(), confirm: vi.fn() },
}));

import ResourceView from "./ResourceView.vue";

describe("ResourceView reconcile flow", () => {
  it("loads fake batches and requests fake difference details", async () => {
    const batch = { id: 91, batchNo: "FAKE-REC-91", differenceCount: 1 };
    const difference = {
      id: 501,
      differenceNo: "FAKE-DIFF-501",
      localValue: "10",
      externalValue: "9",
    };
    api.reconcileBatches.mockResolvedValue({
      data: [batch],
      page: { total: 1, nextCursor: 0, hasMore: false },
    });
    api.reconcileDetails.mockResolvedValue({
      data: [difference],
      page: { total: 1, nextCursor: 0, hasMore: false },
    });

    const wrapper = mount(ResourceView, {
      props: { resource: "reconcile" },
      global: {
		directives: { loading: {} },
        stubs: {
          "el-button": { template: "<button @click=\"$emit('click')\"><slot /></button>" },
          "el-input": true,
          "el-input-number": true,
          "el-select": true,
          "el-option": true,
          "el-form": true,
          "el-form-item": true,
          "el-dialog": true,
          "el-table": true,
          "el-table-column": true,
          "el-descriptions": true,
          "el-descriptions-item": true,
          ListPager: true,
        },
      },
    });
    await flushPromises();

    expect(api.reconcileBatches).toHaveBeenCalledOnce();
    await (wrapper.vm as unknown as { showDetail: (row: typeof batch) => Promise<void> }).showDetail(batch);
    await flushPromises();
    expect(api.reconcileDetails).toHaveBeenCalledWith(91, { limit: 100 });
  });
});
