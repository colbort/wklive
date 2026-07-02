export const chatEventType = {
  WS_CONNECTED: 1,
  MESSAGE: 2,
  SYSTEM_NOTICE: 3,
  USER_JOIN: 4,
  USER_LEAVE: 5,
  QUEUE_UPDATE: 6,
  AGENT_JOIN: 7,
  AGENT_ACCEPTED: 8,
  AGENT_LEAVE: 9,
  SESSION_CLOSE: 13,
  EVALUATION_SUBMIT: 15,
  TYPING: 16,
  MESSAGE_DELIVERED: 17,
  MESSAGE_RECALL: 19,
  MESSAGE_DELETE: 20,
  ERROR: 22,
} as const;

export type ChatEventType = (typeof chatEventType)[keyof typeof chatEventType];

export const chatAdminWsEventTypes: ReadonlySet<ChatEventType> = new Set(
  Object.values(chatEventType),
);
