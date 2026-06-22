import { computed, ref } from "vue";
import {
  chatWsEvents,
  createChatSocket,
  sendChatSocketUserMessage,
} from "@/api/chat";
import type {
  ChatMessage,
  ChatMessageResp,
  ChatWsEvent,
  ConnectedPayload,
  SendUserMessagePayload,
} from "@/types/chat";

interface ConnectOptions {
  merchantId: number;
  token: string;
}

const reconnectDelays = [1000, 2000, 5000, 10000, 15000];

export function useChatSocket() {
  const socket = ref<WebSocket | null>(null);
  const connected = ref<ConnectedPayload | null>(null);
  const messages = ref<ChatMessage[]>([]);
  const error = ref("");
  const reconnectingIn = ref(0);
  const status = ref<"idle" | "connecting" | "open" | "closed" | "reconnecting">(
    "idle",
  );

  let lastOptions: ConnectOptions | null = null;
  let manualClose = false;
  let suppressNextCloseReconnect = false;
  let reconnectAttempts = 0;
  let reconnectTimer: ReturnType<typeof window.setTimeout> | null = null;
  let reconnectTicker: ReturnType<typeof window.setInterval> | null = null;

  const isOpen = computed(() => status.value === "open");
  const isTemporary = computed(() => Boolean(connected.value?.temporary));
  const reconnectLabel = computed(() => {
    if (status.value !== "reconnecting" || reconnectingIn.value <= 0) {
      return "";
    }
    return `${reconnectingIn.value}s 后重连`;
  });

  function connect(merchantId: number, token = "") {
    manualClose = false;
    lastOptions = { merchantId, token };
    reconnectAttempts = 0;
    clearReconnectTimer();
    openSocket(lastOptions);
  }

  function openSocket(options: ConnectOptions) {
    closeSocketOnly(true);
    error.value = "";
    reconnectingIn.value = 0;
    status.value = reconnectAttempts > 0 ? "reconnecting" : "connecting";
    const ws = createChatSocket({
      merchantId: options.merchantId,
      token: options.token,
      onOpen: () => {
        reconnectAttempts = 0;
        reconnectingIn.value = 0;
        status.value = "open";
      },
      onEvent: (event) => {
        handleSocketMessage(event.data);
      },
      onError: () => {
        error.value = "连接异常";
      },
      onClose: () => {
        if (socket.value === ws) {
          socket.value = null;
          status.value = "closed";
        }
        if (suppressNextCloseReconnect) {
          suppressNextCloseReconnect = false;
          return;
        }
        scheduleReconnect();
      },
    });
    socket.value = ws;
  }

  function sendText(content: string, nickname: string) {
    const text = content.trim();
    if (!socket.value || socket.value.readyState !== WebSocket.OPEN || !text) {
      return;
    }
    const payload: SendUserMessagePayload = {
      messageType: 1,
      content: text,
      senderNickname: nickname.trim(),
    };
    sendChatSocketUserMessage(socket.value, payload);
  }

  function close() {
    manualClose = true;
    lastOptions = null;
    clearReconnectTimer();
    closeSocketOnly(false);
    connected.value = null;
    status.value = "idle";
  }

  function resetMessages() {
    messages.value = [];
  }

  function closeSocketOnly(suppressReconnect: boolean) {
    if (socket.value) {
      suppressNextCloseReconnect = suppressReconnect;
      socket.value.close();
    }
    socket.value = null;
  }

  function scheduleReconnect() {
    if (manualClose || !lastOptions) {
      return;
    }
    clearReconnectTimer();

    const delay =
      reconnectDelays[Math.min(reconnectAttempts, reconnectDelays.length - 1)];
    reconnectAttempts += 1;
    reconnectingIn.value = Math.ceil(delay / 1000);
    status.value = "reconnecting";

    reconnectTicker = window.setInterval(() => {
      reconnectingIn.value = Math.max(0, reconnectingIn.value - 1);
    }, 1000);
    reconnectTimer = window.setTimeout(() => {
      clearReconnectTimer();
      if (lastOptions && !manualClose) {
        openSocket(lastOptions);
      }
    }, delay);
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      window.clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (reconnectTicker) {
      window.clearInterval(reconnectTicker);
      reconnectTicker = null;
    }
    reconnectingIn.value = 0;
  }

  function handleSocketMessage(payload: string) {
    try {
      handleEvent(JSON.parse(payload) as ChatWsEvent);
    } catch {
      error.value = "无效的消息格式";
    }
  }

  function handleEvent(event: ChatWsEvent) {
    if (event.type === chatWsEvents.connected) {
      connected.value = event.data as ConnectedPayload;
      return;
    }
    if (event.type === chatWsEvents.error) {
      const data = event.data as { message?: string };
      error.value = data.message || "消息发送失败";
      return;
    }
    if (event.type === chatWsEvents.sendUserMessageResult) {
      const resp = event.data as ChatMessageResp;
      if (resp.code !== 200) {
        error.value = resp.msg || "消息发送失败";
        return;
      }
      pushMessage(resp.data);
      return;
    }
    if (event.type === chatWsEvents.message) {
      pushMessage(event.data as ChatMessage);
    }
  }

  function pushMessage(message: ChatMessage) {
    if (!message?.messageNo) {
      return;
    }
    if (messages.value.some((item) => item.messageNo === message.messageNo)) {
      return;
    }
    messages.value.push(message);
  }

  return {
    connected,
    error,
    isOpen,
    isTemporary,
    messages,
    reconnectLabel,
    reconnectingIn,
    status,
    close,
    connect,
    resetMessages,
    sendText,
  };
}
