import { defineComponent, h, nextTick } from "vue";
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import WorkbenchChatPanel from "./WorkbenchChatPanel.vue";

const ElInputStub = defineComponent({
  props: { modelValue: { type: String, default: "" } },
  emits: ["update:modelValue", "input"],
  setup(props, { emit }) {
    return () =>
      h("textarea", {
        value: props.modelValue,
        onInput: (event: Event) => {
          const value = (event.target as HTMLTextAreaElement).value;
          emit("update:modelValue", value);
          emit("input", value);
        },
      });
  },
});

describe("WorkbenchChatPanel typing", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("emits debounced typing and a stop event using fake input data", async () => {
    vi.useFakeTimers();
    const wrapper = mount(WorkbenchChatPanel, {
      props: {
        session: { sessionNo: "fake-session" } as never,
        messages: [],
        loading: false,
        activeNeedsAccept: false,
        activeClosed: false,
        canReply: true,
        canAccept: true,
        acceptDisabledReason: "",
        wsOnline: true,
        agentId: 7,
        userId: 8,
        showGuestRefreshNotice: false,
      },
      global: {
        stubs: {
          "el-input": ElInputStub,
          "el-button": { template: "<button><slot /></button>" },
		  "el-tooltip": { template: "<span><slot /></span>" },
          ChatMessageBubble: true,
        },
      },
    });

    await wrapper.get("textarea").setValue("fake reply");
    vi.advanceTimersByTime(250);
    await nextTick();
    expect(wrapper.emitted("typing")?.at(-1)).toEqual(["fake reply"]);

    vi.advanceTimersByTime(1250);
    await nextTick();
    expect(wrapper.emitted("typing")?.at(-1)).toEqual([""]);
  });
});
