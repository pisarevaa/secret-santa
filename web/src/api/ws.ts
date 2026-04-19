type MessageHandler = (msg: {
  type: string;
  id?: number;
  role?: string;
  from_me?: boolean;
  body?: string;
  created_at?: string;
  reason?: string;
}) => void;

export class ChatSocket {
  private ws: WebSocket | null = null;
  private url: string;
  private onMessage: MessageHandler;
  private reconnectAttempts = 0;
  private maxReconnectDelay = 30000;
  private closed = false;

  constructor(groupId: number, onMessage: MessageHandler) {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    this.url = `${protocol}//${window.location.host}/ws/groups/${groupId}`;
    this.onMessage = onMessage;
    this.connect();
  }

  private connect() {
    if (this.closed) return;

    this.ws = new WebSocket(this.url);

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        this.onMessage(msg);
      } catch {
        // ignore parse errors
      }
    };

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };

    this.ws.onclose = () => {
      if (this.closed) return;
      this.scheduleReconnect();
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  private scheduleReconnect() {
    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts) + Math.random() * 1000,
      this.maxReconnectDelay
    );
    this.reconnectAttempts++;
    setTimeout(() => this.connect(), delay);
  }

  send(role: "santa" | "recipient", body: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: "send", role, body }));
    }
  }

  close() {
    this.closed = true;
    this.ws?.close();
  }
}
