import { cp, mkdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const currentDir = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(currentDir, "..");
const source = path.join(rootDir, "lib", "widget", "sdk-template.js");
const targetDir = path.join(rootDir, "public", "sdk");
const target = path.join(targetDir, "cs-agent-widget.js");

await mkdir(targetDir, { recursive: true });
await cp(source, target);

console.log(`sdk written to ${target}`);
