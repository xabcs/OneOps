<script setup lang="ts">
    import { onMounted, onUnmounted, ref, nextTick } from "vue";
    import { Terminal } from "xterm";
    import { FitAddon } from "xterm-addon-fit";
    import { WebLinksAddon } from "xterm-addon-web-links";
    import "xterm/css/xterm.css";

    interface Props {
        sessionId: number;
        serverId: number;
        serverName: string;
        serverIp: string;
        loginAccount: string;
        websocketUrl?: string;
    }

    const props = defineProps<Props>();

    let terminal: Terminal | null = null;
    let fitAddon: FitAddon | null = null;
    let ws: WebSocket | null = null;
    const terminalRef = ref<HTMLDivElement>();

    // 防抖定时器
    let resizeTimeout: ReturnType<typeof setTimeout> | null = null;
    let pendingFit = false;

    // 获取 token
    function getToken(): string {
        let token =
            localStorage.getItem("SOY_token") ||
            localStorage.getItem("token") ||
            "";

        // 移除可能存在的引号（存储时可能被包裹）
        token = token.replace(/^["']|["']$/g, "");

        return token;
    }

    // 初始化终端
    async function initTerminal() {
        if (!terminalRef.value) {
            console.error("terminalRef.value 不存在！");
            return;
        }

        // 创建终端实例
        terminal = new Terminal({
            cursorBlink: true,
            fontSize: 14,
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            theme: {
                background: "#1e1e1e",
                foreground: "#cccccc",
                cursor: "#cccccc",
                selection: "rgba(255, 255, 255, 0.3)",
                black: "#000000",
                red: "#cd3131",
                green: "#0dbc79",
                yellow: "#e5e510",
                blue: "#2472c8",
                magenta: "#bc3fbc",
                cyan: "#11a8cd",
                white: "#e5e5e5",
                brightBlack: "#666666",
                brightRed: "#f14c4c",
                brightGreen: "#23d18b",
                brightYellow: "#f5f543",
                brightBlue: "#3b8eea",
                brightMagenta: "#d670d6",
                brightCyan: "#29b8db",
                brightWhite: "#ffffff",
            },
            scrollback: 10000,
            tabStopWidth: 8,
        });

        // 添加插件
        fitAddon = new FitAddon();
        terminal.loadAddon(fitAddon);
        terminal.loadAddon(new WebLinksAddon());

        // 挂载终端
        terminal.open(terminalRef.value);

        // 等待 DOM 渲染完成后再执行 fit
        await nextTick();
        requestAnimationFrame(() => {
            fitAddon?.fit();
        });

        // 欢迎信息
        terminal.writeln(
            `\x1b[1;36m${props.serverName}\x1b[0m (${props.serverIp})`
        );
        terminal.writeln(`\x1b[1;34m登录用户: ${props.loginAccount}\x1b[0m`);
        terminal.writeln("");

        // 监听用户输入
        terminal.onData((data) => {
            if (ws && ws.readyState === WebSocket.OPEN) {
                ws.send(
                    JSON.stringify({
                        type: "input",
                        data: data,
                    })
                );
            }
        });

        // 建立 WebSocket 连接
        connectWebSocket();
    }

    // 获取后端服务地址
    function getBackendHost(): string {
        // 从环境变量获取后端服务地址
        const baseUrl =
            import.meta.env.VITE_SERVICE_BASE_URL || "http://localhost:8082/api";
        const url = new URL(baseUrl);
        return url.host;
    }

    // 连接 WebSocket
    function connectWebSocket() {
        const token = getToken();
        const backendHost = getBackendHost();

        let fullWsUrl: string;

        // 如果提供了后端返回的 websocketUrl，使用它
        if (props.websocketUrl) {
            // 后端返回的是相对路径，需要添加协议和后端主机
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            fullWsUrl = `${protocol}//${backendHost}${props.websocketUrl}?token=${token}`;
        } else {
            // 否则使用默认方式构建 URL
            const wsUrl = `/api/cmdb/sessions/${props.sessionId}/ws?token=${token}`;
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            fullWsUrl = `${protocol}//${backendHost}${wsUrl}`;
        }

        console.log("WebSocket 连接:", fullWsUrl);

        ws = new WebSocket(fullWsUrl);

        ws.onopen = () => {
            console.log("WebSocket 连接成功");
        };

        ws.onmessage = (event) => {
            if (terminal) {
                terminal.write(event.data);
            }
        };

        ws.onerror = (error) => {
            console.error("WebSocket 错误:", error);
            if (terminal) {
                terminal.writeln(`\x1b[1;31m✗ 连接错误\x1b[0m`);
            }
        };

        ws.onclose = (event) => {
            console.log("WebSocket 关闭:", event.code, event.reason);
            if (terminal) {
                terminal.writeln(
                    `\r\n\x1b[1;31m✗ 连接已关闭 (code: ${event.code})\x1b[0m`
                );
            }
        };
    }

    // 延迟执行 fit，避免频繁调用
    function scheduleFit() {
        if (pendingFit) return;
        pendingFit = true;

        // 使用较长的延迟，确保容器尺寸稳定后再执行
        if (resizeTimeout) {
            clearTimeout(resizeTimeout);
        }

        resizeTimeout = setTimeout(() => {
            if (fitAddon && terminal) {
                try {
                    fitAddon.fit();
                } catch (error) {
                    console.warn("fit error:", error);
                }
            }
            pendingFit = false;
        }, 300);
    }

    // 窗口 resize 处理
    function handleResize() {
        scheduleFit();
    }

    // 处理可见性变化（切换标签时）
    function handleVisibilityChange() {
        if (terminalRef.value) {
            const style = window.getComputedStyle(terminalRef.value);
            if (style.display !== "none" && fitAddon) {
                // 当元素变为可见时，重新 fit 终端
                scheduleFit();
            }
        }
    }

    onMounted(() => {
        initTerminal();

        // 添加窗口 resize 监听
        window.addEventListener("resize", handleResize);

        // 使用 MutationObserver 监听容器 display 属性变化
        const observer = new MutationObserver((mutations) => {
            for (const mutation of mutations) {
                if (
                    mutation.type === "attributes" &&
                    mutation.attributeName === "style"
                ) {
                    handleVisibilityChange();
                }
            }
        });

        if (terminalRef.value) {
            observer.observe(terminalRef.value, {
                attributes: true,
                attributeFilter: ["style"],
            });
        }
    });

    onUnmounted(() => {
        // 清理定时器
        if (resizeTimeout) {
            clearTimeout(resizeTimeout);
        }

        // 移除事件监听
        window.removeEventListener("resize", handleResize);

        // 关闭 WebSocket
        if (ws) {
            ws.close();
            ws = null;
        }

        // 销毁终端
        if (terminal) {
            terminal.dispose();
            terminal = null;
        }
    });
</script>

<template>
    <div ref="terminalRef" class="xterm-terminal"></div>
</template>

<style lang="scss">
    .xterm-terminal {
        width: 100%;
        height: 100%;
        display: flex;
        flex-direction: column;

        :deep(.xterm) {
            width: 100%;
            height: 100%;
            background-color: #1e1e1e;
        }

        :deep(.xterm-viewport) {
            background-color: #1e1e1e;
            scrollbar-width: thin;
            scrollbar-color: #444 #1e1e1e;
        }

        :deep(.xterm-viewport::-webkit-scrollbar) {
            width: 8px;
            height: 8px;
        }

        :deep(.xterm-viewport::-webkit-scrollbar-track) {
            background: #1e1e1e;
        }

        :deep(.xterm-viewport::-webkit-scrollbar-thumb) {
            background: #444;
            border-radius: 4px;

            &:hover {
                background: #555;
            }
        }
    }
</style>
