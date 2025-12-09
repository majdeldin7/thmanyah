const express = require("express");
const axios = require("axios");
const { Client } = require("pg");
const client = require("prom-client");

const app = express();

// ---- Prometheus Metrics Setup ---- //
const collectDefaultMetrics = client.collectDefaultMetrics;
collectDefaultMetrics();

const httpRequestCounter = new client.Counter({
  name: "http_requests_total",
  help: "Total number of HTTP requests",
  labelNames: ["method", "route", "status"],
});

const httpRequestDuration = new client.Histogram({
  name: "http_request_duration_seconds",
  help: "Duration of HTTP requests in seconds",
  labelNames: ["method", "route", "status"],
  buckets: [0.1, 0.5, 1, 2, 5],
});

// ---- Metrics Middleware ---- //
app.use((req, res, next) => {
  const start = process.hrtime();
  res.on("finish", () => {
    const diff = process.hrtime(start);
    const duration = diff[0] + diff[1] / 1e9;

    httpRequestCounter.inc({
      method: req.method,
      route: req.path,
      status: res.statusCode
    });

    httpRequestDuration.observe(
      { method: req.method, route: req.path, status: res.statusCode },
      duration
    );
  });

  next();
});

// ---- Routes ---- //
app.get("/health", (req, res) => {
  res.json({ service: "api-service", status: "ok" });
});

app.get("/auth-check", async (req, res) => {
  try {
    const auth = await axios.get("http://auth-service.auth-service/verify");
    res.json({ ok: true, auth: auth.data });
  } catch (err) {
    res.status(500).json({ ok: false, error: err.message });
  }
});

app.get("/image-check", async (req, res) => {
  try {
    const img = await axios.get("http://image-service.image-service/info");
    res.json({ ok: true, image: img.data });
  } catch (err) {
    res.status(500).json({ ok: false, error: err.message });
  }
});

app.get("/db-check", async (req, res) => {
  try {
    const result = await global.pgClient.query("SELECT NOW()");
    res.json({ connected: true, time: result.rows[0].now });
  } catch (err) {
    res.status(500).json({ connected: false, error: err.message });
  }
});

// ---- /metrics Endpoint ---- //
app.get("/metrics", async (req, res) => {
  res.set("Content-Type", client.register.contentType);
  res.end(await client.register.metrics());
});

// ---- Heartbeat With Timeout ---- //
async function heartbeatQuery() {
  return Promise.race([
    global.pgClient.query("SELECT 1"),
    new Promise((_, reject) =>
      setTimeout(() => reject(new Error("DB heartbeat timeout")), 3000)
    )
  ]);
}

// ---- DB Heartbeat Loop ---- //
function startDBHeartbeat() {
  setInterval(async () => {
    try {
      await heartbeatQuery();
      console.log("DB heartbeat OK");
    } catch (err) {
      console.error("❌ DB heartbeat failed:", err.message);
      process.exit(1);
    }
  }, 5000);
}

// ---- Start Server AFTER DB Connect ---- //
async function startServer() {
  const pgClient = new Client({
    host: process.env.DB_HOST,
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_NAME
  });

  try {
    console.log("Checking DB connection...");
    await pgClient.connect();
    console.log("Database connected successfully.");

    pgClient.on("error", (err) => {
      console.error("❌ PostgreSQL connection error detected!", err);
      process.exit(1);
    });

    global.pgClient = pgClient;

    startDBHeartbeat();

    app.listen(3000, () => console.log("api-service running on port 3000"));
  } catch (err) {
    console.error("❌ Failed to connect to the database:", err.message);
    process.exit(1);
  }
}

startServer();