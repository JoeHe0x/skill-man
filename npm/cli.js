#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const root = path.join(__dirname, "..");
const binName = process.platform === "win32" ? "skill-man.exe" : "skill-man";
const binPath = path.join(root, "dist", binName);

if (!fs.existsSync(binPath)) {
  console.error(
    "skill-man binary not found. Reinstall:\n  npm install -g skill-man\n  npm rebuild -g skill-man"
  );
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);
