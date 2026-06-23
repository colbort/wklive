import type { OptionItem } from "@/api/chat";

export type DisplayOptionItem = OptionItem & { label: string };

const optionLabelFallbacks: Record<string, string> = {
  "chat.agent.status.offline": "离线",
  "chat.agent.status.online": "在线",
  "chat.agent.status.busy": "忙碌",
  "chat.agent.status.resting": "休息",
};

export function withOptionLabels(items: OptionItem[]): DisplayOptionItem[] {
  return items.map((item) => ({
    ...item,
    label: item.label || optionLabelFallbacks[item.key] || item.key,
  }));
}
