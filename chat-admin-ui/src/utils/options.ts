import type { OptionGroup, OptionItem } from "@/api/chat";

export type DisplayOptionItem = OptionItem & {
  label: string;
  tagType?: "success" | "info" | "warning" | "danger" | "primary";
};

const optionLabelFallbacks: Record<string, string> = {
  CHAT_AGENT_STATUS_OFFLINE: "离线",
  CHAT_AGENT_STATUS_ONLINE: "在线",
  CHAT_AGENT_STATUS_BUSY: "忙碌",
  CHAT_AGENT_STATUS_RESTING: "休息",
  CHAT_SESSION_STATUS_WAITING: "待接待",
  CHAT_SESSION_STATUS_SERVING: "服务中",
  CHAT_SESSION_STATUS_PENDING_USER: "等待用户回复",
  CHAT_SESSION_STATUS_PENDING_AGENT: "等待客服回复",
  CHAT_SESSION_STATUS_CLOSED: "已结束",
  CHAT_SENDER_TYPE_USER: "用户",
  CHAT_SENDER_TYPE_AGENT: "客服",
  CHAT_SENDER_TYPE_SYSTEM: "系统",
  CHAT_MESSAGE_TYPE_TEXT: "文本",
  CHAT_MESSAGE_TYPE_IMAGE: "图片",
  CHAT_MESSAGE_TYPE_FILE: "文件",
  CHAT_MESSAGE_TYPE_ORDER: "订单",
  CHAT_MESSAGE_TYPE_SYSTEM: "系统提示",
};

const optionTagFallbacks: Record<string, DisplayOptionItem["tagType"]> = {
  CHAT_AGENT_STATUS_OFFLINE: "info",
  CHAT_AGENT_STATUS_ONLINE: "success",
  CHAT_AGENT_STATUS_BUSY: "warning",
  CHAT_AGENT_STATUS_RESTING: "info",
};

export function withOptionLabels(items: OptionItem[]): DisplayOptionItem[] {
  return items.map((item) => ({
    ...item,
    label: item.label || optionLabelFallbacks[item.code] || item.code,
    tagType: item.tagType || optionTagFallbacks[item.code],
  }));
}

export function optionGroup(groups: OptionGroup[] | undefined, key: string) {
  return groups?.find((group) => group.key === key)?.options || [];
}
