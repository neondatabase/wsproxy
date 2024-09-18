import { Pool, neonConfig } from '@neondatabase/serverless';
import { drizzle } from 'drizzle-orm/neon-serverless';
import { sql } from 'drizzle-orm';

// Docs
// https://orm.drizzle.team/docs/get-started-postgresql#neon-postgres
// Test with
// NEON_WS_PROXY_HOST=$(docker compose port neon 80) bun run serverless.ts

const NEON_WS_PROXY_HOST = process.env.NEON_WS_PROXY_HOST;
console.log('NEON_WS_PROXY_HOST', NEON_WS_PROXY_HOST);

if (!NEON_WS_PROXY_HOST) {
    throw new Error('NEON_WS_PROXY_HOST is not set');
}

// Set the WebSocket proxy to work with the local instance
neonConfig.wsProxy = () => `${process.env.NEON_WS_PROXY_HOST}/v1`;

// Disable TLS when running on local machine
neonConfig.useSecureWebSocket = false;
// Disable all authentication and encryption
neonConfig.pipelineTLS = false;
neonConfig.pipelineConnect = false;

const connectionString = `pgsql://postgres:neon-password@placeholder/neon-db`;

const pool = new Pool({ connectionString });
const db = drizzle(pool);

const result = await db.execute(sql`SELECT version()`);
const res = result.rows[0];

console.log(res);

await pool.end();
process.exit(0);
