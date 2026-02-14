import { Elysia } from "elysia";
import { config, validateConfig } from "./config";
import { runMigrations } from "./db/migrate";
import { closePool } from "./db/connection";
import { bugReportRoutes } from "./routes/bugReports";
import { adminRoutes } from "./routes/admin";

// Validate environment
validateConfig();

// Run database migrations
await runMigrations();

const app = new Elysia()
  .onError(({ error, set }) => {
    console.error("Unhandled error:", error);
    set.status = 500;
    return { success: false, message: "Internal server error" };
  })
  .get("/health", () => ({ status: "ok", timestamp: new Date().toISOString() }))
  .use(bugReportRoutes)
  .use(adminRoutes)
  .listen({
    port: config.port,
    maxBody: 50 * 1024 * 1024, // 50MB
  });

console.log(
  `EMLy Bug Report API running on http://localhost:${app.server?.port}`
);

// Graceful shutdown
process.on("SIGINT", async () => {
  console.log("Shutting down...");
  await closePool();
  process.exit(0);
});

process.on("SIGTERM", async () => {
  console.log("Shutting down...");
  await closePool();
  process.exit(0);
});
