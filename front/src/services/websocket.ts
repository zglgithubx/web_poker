export class WebSocketService {
  private ws: WebSocket | null = null;
  private messageHandlers: Map<string, (data: any) => void> = new Map();

  constructor(private url: string) {}

  connect() {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      
      console.log('WebSocket 连接已建立');
      // 发送初始化消息
      this.send({
        type: 'init',
        data: {
          playerId: localStorage.getItem('playerId') || crypto.randomUUID(),
          playerName: localStorage.getItem('playerName') || '玩家' + Math.floor(Math.random() * 1000)
        }
      });
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      const handler = this.messageHandlers.get(message.type);
      if (handler) {
        handler(message.data);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket 错误:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket 连接已关闭');
    };
  }

  send(message: { type: string; data: any }) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  on(type: string, handler: (data: any) => void) {
    this.messageHandlers.set(type, handler);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
} 