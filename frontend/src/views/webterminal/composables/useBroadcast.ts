export interface BroadcastMessage {
  type: string;
  data?: any;
  timestamp?: number;
}

export function useBroadcast(channelName: string) {
  let channel: BroadcastChannel | null = null;

  try {
    channel = new BroadcastChannel(channelName);
  } catch (error) {
    console.warn('BroadcastChannel not supported:', error);
  }

  function broadcast(type: string, data?: any): void {
    if (channel) {
      const message: BroadcastMessage = {
        type,
        data,
        timestamp: Date.now()
      };
      channel.postMessage(message);
    }
  }

  function onMessage(handler: (message: BroadcastMessage) => void): void {
    if (channel) {
      channel.onmessage = event => {
        handler(event.data);
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
