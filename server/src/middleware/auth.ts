import { Elysia } from "elysia";
import { config } from "../config";

export const apiKeyGuard = new Elysia({ name: "api-key-guard" }).derive(
  { as: "scoped" },
  ({ headers, error }) => {
    const key = headers["x-api-key"];
    if (!key || key !== config.apiKey) {
      return error(401, { success: false, message: "Invalid or missing API key" });
    }
    return {};
  }
);

export const adminKeyGuard = new Elysia({ name: "admin-key-guard" }).derive(
  { as: "scoped" },
  ({ headers, error }) => {
    const key = headers["x-admin-key"];
    if (!key || key !== config.adminKey) {
      return error(401, { success: false, message: "Invalid or missing admin key" });
    }
    return {};
  }
);
