export interface BroadcastMessage<T = unknown> {
  type: string;
  data?: T;
  timestamp?: number;
}

export function useBroadcast(channelName: string) {
  let channel: BroadcastChannel | null = null;

  try {
    channel = new BroadcastChannel(channelName);
  } catch (error) {
    console.warn('BroadcastChannel not supported:', error);
  }

  function broadcast<T = unknown>(type: string, data?: T): void {
    if (channel) {
      const message: BroadcastMessage<T> = {
        type,
        data,
        timestamp: Date.now()
      };
      channel.postMessage(message);
    }
  }

  function onMessage<T = unknown>(handler: (message: BroadcastMessage<T>) => void): void {
    if (channel) {
      channel.onmessage = event => {
        handler(event.data as BroadcastMessage<T>);
      };
    }
  }

  function dispose(): void {
    if (channel) {
      channel.close();
      channel = null;
    }
  }

  return {
    broadcast,
    onMessage,
    dispose
  };
}
