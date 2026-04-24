import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const currentDir = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(currentDir, "..");
const source = path.join(rootDir, "lib", "sdk", "cs-ai-agent-sdk.js");
const targetDir = path.join(rootDir, "public", "sdk");
const target = path.join(targetDir, "cs-ai-agent-sdk.min.js");

function minifyJavaScript(sourceText) {
  return sourceText
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/(^|[^:])\/\/.*$/gm, "$1")
    .replace(/\s+/g, " ")
    .replace(/\s*([{}()[\];,:?=+\-*/<>|&])\s*/g, "$1")
    .trim();
}

const sourceText = await readFile(source, "utf8");
const output = `${minifyJavaScript(sourceText)}\n`;

await mkdir(targetDir, { recursive: true });
await writeFile(target, output, "utf8");

console.log(`sdk written to ${target}`);
