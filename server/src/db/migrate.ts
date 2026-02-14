import { readFileSync } from "fs";
import { join } from "path";
import { getPool } from "./connection";

export async function runMigrations(): Promise<void> {
  const pool = getPool();
  const schemaPath = join(import.meta.dir, "schema.sql");
  const schema = readFileSync(schemaPath, "utf-8");

  // Split on semicolons, filter empty statements
  const statements = schema
    .split(";")
    .map((s) => s.trim())
    .filter((s) => s.length > 0);

  for (const statement of statements) {
    await pool.execute(statement);
  }

  // Additive migrations for existing databases
  const alterMigrations = [
    `ALTER TABLE bug_reports ADD COLUMN IF NOT EXISTS hostname VARCHAR(255) NOT NULL DEFAULT '' AFTER hwid`,
    `ALTER TABLE bug_reports ADD COLUMN IF NOT EXISTS os_user VARCHAR(255) NOT NULL DEFAULT '' AFTER hostname`,
    `ALTER TABLE bug_reports ADD INDEX IF NOT EXISTS idx_hostname (hostname)`,
    `ALTER TABLE bug_reports ADD INDEX IF NOT EXISTS idx_os_user (os_user)`,
  ];

  for (const migration of alterMigrations) {
    try {
      await pool.execute(migration);
    } catch {
      // Column/index already exists — safe to ignore
    }
  }

  console.log("Database migrations completed");
}
