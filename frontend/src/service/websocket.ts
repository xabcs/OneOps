import { ElNotification } from 'element-plus';
import { useAuthStore } from '@/store/modules/auth';

interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: string;
}

type MessageHandler = (data: any) => void;

class MonitoringWebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectTimer: number | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private subscriptions: Set<string> = new Set();
  private messageHandlers: Map<string, MessageHandler[]> = new Map();
  private url: string = '';

  constructor() {
    // 延迟初始化，等待首次调用时确保 token 已加载
    // 不在构造函数中直接连接，而是在首次使用时连接
  }

  private connect() {
    const authStore = useAuthStore();
    const token = authStore.token;

    if (!token) {
      console.warn('[WebSocket] No token found, skipping connection');
      return;
    }

    // 从环境变量获取API地址，构建WebSocket URL
    const apiBase = import.meta.env.VITE_SERVICE_BASE_URL || 'http://localhost:8082/api';
    const wsHost = apiBase.replace(/^https?:\/\//, '').replace(/\/api$/, '');
    const wsProtocol = apiBase.startsWith('https') ? 'wss://' : 'ws://';
    const wsUrl = `${wsProtocol}${wsHost}/api/monitoring/ws?token=${token}`;
    this.url = wsUrl;

    console.log('[WebSocket] 连接URL:', wsUrl);
    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('[WebSocket] 连接成功');
        this.reconnectAttempts = 0;

        // 清除重连定时器
        if (this.reconnectTimer) {
          clearTimeout(this.reconnectTimer);
          this.reconnectTimer = null;
        }

        // 重新订阅
        this.resubscribe();
      };

      this.ws.onmessage = event => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('[WebSocket] 消息解析失败:', error);
        }
      };

      this.ws.onerror = error => {
        console.error('[WebSocket] 连接错误:', error);
      };

      this.ws.onclose = () => {
        console.log('[WebSocket] 连接关闭，5秒后重连...');
        this.scheduleReconnect();
      };
    } catch (error) {
      console.error('[WebSocket] 创建连接失败:', error);
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('[WebSocket] 达到最大重连次数，停止重连');
      return;
    }

    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectAttempts++;
      console.log(`[WebSocket] 尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      this.connect();
    }, 5000) as unknown as number;
  }

  private handleMessage(message: WebSocketMessage) {
    const { type, data } = message;
    const handlers = this.messageHandlers.get(type);

    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(data);
        } catch (error) {
          console.error(`[WebSocket] 消息处理器错误 (${type}):`, error);
        }
      });
    }

    // 默认处理：告警消息显示通知
    if (type === 'alert' && !handlers?.length) {
      this.showAlertNotification(data);
    }
  }

  private showAlertNotification(alert: any) {
    const levelMap: Record<string, 'success' | 'warning' | 'error' | 'info'> = {
      critical: 'error',
      high: 'error',
      medium: 'warning',
      low: 'info',
      info: 'info'
    };

    ElNotification({
      title: `${alert.level === 'critical' ? '🚨' : '⚠️'} ${alert.hostname}`,
      message: alert.message,
      type: levelMap[alert.level] || 'info',
      duration: 0,
      onClick: () => {
        // 跳转到告警详情页面
        window.location.hash = `/monitoring/alerts`;
      }
    });
  }

  // 订阅数据更新
  subscribe(type: string, handler: MessageHandler) {
    // 确保已连接（首次订阅时建立连接）
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED) {
      this.connect();
    }

    this.subscriptions.add(type);

    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, []);
    }

    this.messageHandlers.get(type)!.push(handler);

    // 发送订阅消息
    this.send({
      action: 'subscribe',
      topic: type
    });
  }

  // 取消订阅
  unsubscribe(type: string, handler?: MessageHandler) {
    if (handler) {
      const handlers = this.messageHandlers.get(type);
      if (handlers) {
        const index = handlers.indexOf(handler);
        if (index > -1) {
          handlers.splice(index, 1);
        }
      }
    } else {
      this.messageHandlers.delete(type);
      this.subscriptions.delete(type);
    }

    // 发送取消订阅消息
    this.send({
      action: 'unsubscribe',
      topic: type
    });
  }

  // 重新订阅
  private resubscribe() {
    this.subscriptions.forEach(type => {
      this.send({
        action: 'subscribe',
        topic: type
      });
    });
  }

  // 发送消息
  private send(message: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  // 关闭连接
  close() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }
    if (this.ws) {
      this.ws.close();
    }
  }

  // 获取连接状态
  get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }

  // 是否已连接
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// 单例实例
let wsClient: MonitoringWebSocketClient | null = null;

export function useWebSocket() {
  if (!wsClient) {
    wsClient = new MonitoringWebSocketClient();
  }

  return {
    client: wsClient,
    subscribe: wsClient.subscribe.bind(wsClient),
    unsubscribe: wsClient.unsubscribe.bind(wsClient),
    isConnected: () => wsClient.isConnected,
    close: () => wsClient?.close()
  };
}
