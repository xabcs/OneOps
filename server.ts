import express from "express";
import { createServer as createViteServer } from "vite";
import path from "path";
import mysql from "mysql2/promise";

async function startServer() {
  console.log(">>> NexOps Server Starting <<<");
  const app = express();
  const PORT = 3000;

  app.use(express.json());

  app.use((req, res, next) => {
    console.log(`[${new Date().toISOString()}] ${req.method} ${req.url}`);
    next();
  });

  // --- Database Connection (Non-blocking) ---
  let db: mysql.Connection | null = null;
  const connectDB = async () => {
    try {
      db = await mysql.createConnection({
        host: "60.191.116.75",
        port: 38089,
        user: "root",
        password: "123456",
        database: "ops",
        connectTimeout: 5000 // 5 seconds timeout
      });
      console.log("Connected to MySQL database at 60.191.116.75:38089");
    } catch (err) {
      console.error("Failed to connect to MySQL database:", err);
      console.log("Falling back to mock data...");
    }
  };
  connectDB();

  // --- API Routes ---

  app.get("/api/health", (req, res) => {
    res.json({ status: "ok", dbConnected: !!db });
  });

  app.post("/api/login", async (req, res) => {
    const { username, password } = req.body;
    
    if (db) {
      try {
        const [rows]: any = await db.execute(
          "SELECT id, username, password FROM users WHERE username = ? AND password = ?",
          [username, password]
        );
        if (rows.length > 0) {
          return res.json({
            success: true,
            message: "登录成功",
            token: "db-token-" + Date.now(),
            user: {
              id: rows[0].id,
              username: rows[0].username
            }
          });
        }
      } catch (err) {
        console.error("Login DB error:", err);
      }
    }

    // Fallback or Mock
    if ((username === "admin" && password === "admin") || (username === "admin" && password === "123456")) {
      res.json({
        success: true,
        message: "登录成功 (Mock)",
        token: "mock-token-123456",
        user: { id: 1, username: "admin" }
      });
    } else {
      res.status(401).json({ 
        success: false, 
        message: "用户名或密码错误" 
      });
    }
  });

  app.post("/api/register", (req, res) => {
    res.json({
      success: true,
      message: "注册成功"
    });
  });

  app.get("/api/servers", (req, res) => {
    res.json([
      { id: 1, name: "Production-01", ip: "192.168.1.100", status: "online", cpu: 45, memory: 60 },
      { id: 2, name: "Staging-01", ip: "192.168.1.101", status: "online", cpu: 20, memory: 30 },
      { id: 3, name: "Backup-01", ip: "192.168.1.102", status: "offline", cpu: 0, memory: 0 }
    ]);
  });

  app.get("/api/tasks", (req, res) => {
    res.json([
      { id: 1, name: "Backup Database", status: "completed", time: "2024-01-01 14:15:00" },
      { id: 2, name: "System Update", status: "running", time: "2024-01-01 13:45:00" },
      { id: 3, name: "Log Cleanup", status: "pending", time: "2024-01-01 15:00:00" }
    ]);
  });

  app.get("/api/monitoring", (req, res) => {
    res.json({
      cpu: 45,
      memory: 60,
      disk: 75,
      alerts: [
        { id: 1, level: "critical", message: "CPU usage > 80%", time: "2024-01-01 13:30:00" }
      ]
    });
  });

  // --- Vite Middleware ---

  if (process.env.NODE_ENV !== "production") {
    const vite = await createViteServer({
      configFile: path.resolve(process.cwd(), "frontend/vite.config.ts"),
      server: { 
        middlewareMode: true,
        hmr: process.env.DISABLE_HMR !== 'true'
      },
      appType: "spa",
      root: path.resolve(process.cwd(), "frontend")
    });
    app.use(vite.middlewares);
  } else {
    const distPath = path.join(process.cwd(), "frontend/dist");
    app.use(express.static(distPath));
    app.get("*", (req, res) => {
      res.sendFile(path.join(distPath, "index.html"));
    });
  }

  app.listen(PORT, "0.0.0.0", () => {
    console.log(`Server running on http://localhost:${PORT}`);
    console.log(`Environment: ${process.env.NODE_ENV || 'development'}`);
  });
}

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

process.on('uncaughtException', (err) => {
  console.error('Uncaught Exception:', err);
});

startServer().catch((err) => {
  console.error("Error starting server:", err);
});
