export const chatEventType = {
  WS_CONNECTED: 1,
  MESSAGE: 2,
  SYSTEM_NOTICE: 3,
  QUEUE_UPDATE: 6,
  AGENT_ACCEPTED: 8,
  SESSION_CLOSE: 13,
  EVALUATION_INVITE: 14,
  EVALUATION_SUBMIT: 15,
  TYPING: 16,
  MESSAGE_DELIVERED: 17,
  ERROR: 22,
} as const;

export type ChatEventType = (typeof chatEventType)[keyof typeof chatEventType];

export const chatUiWsEventTypes: ReadonlySet<ChatEventType> = new Set(
  Object.values(chatEventType),
);
