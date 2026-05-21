#!/usr/bin/env node
"use strict";

const { execFileSync, spawnSync } = require("child_process");
const fs = require("fs");
const https = require("https");
const os = require("os");
const path = require("path");

const REPO = "JoeHe0x/skill-man";
const MODULE = "github.com/JoeHe0x/skill-man/cmd/skill-man";

const root = path.join(__dirname, "..");
const { version } = require(path.join(root, "package.json"));
const tag = version.startsWith("v") ? version : `v${version}`;

const installDir = path.join(root, "dist");
const binName = process.platform === "win32" ? "skill-man.exe" : "skill-man";
const binPath = path.join(installDir, binName);

function platformArch() {
  const platform = process.env.npm_config_platform || os.platform();
  const arch = process.env.npm_config_arch || os.arch();
  const goos =
    platform === "win32"
      ? "windows"
      : platform === "darwin"
        ? "darwin"
        : "linux";
  const goarch =
    arch === "x64" || arch === "amd64"
      ? "amd64"
      : arch === "arm64"
        ? "arm64"
        : arch;
  return { goos, goarch };
}

function assetName(ver, goos, goarch) {
  const v = ver.replace(/^v/, "");
  const ext = goos === "windows" ? "zip" : "tar.gz";
  return `skill-man_${v}_${goos}_${goarch}.${ext}`;
}

function download(url) {
  return new Promise((resolve, reject) => {
    const req = https.get(
      url,
      { headers: { "User-Agent": "skill-man-npm-install" } },
      (res) => {
        if (
          res.statusCode &&
          res.statusCode >= 300 &&
          res.statusCode < 400 &&
          res.headers.location
        ) {
          res.resume();
          download(res.headers.location).then(resolve, reject);
          return;
        }
        if (res.statusCode !== 200) {
          res.resume();
          reject(new Error(`HTTP ${res.statusCode} for ${url}`));
          return;
        }
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => resolve(Buffer.concat(chunks)));
      }
    );
    req.on("error", reject);
  });
}

function extractArchive(archivePath, goos) {
  fs.mkdirSync(installDir, { recursive: true });
  if (goos === "windows") {
    execFileSync(
      "tar",
      ["-xf", archivePath, "-C", installDir],
      { stdio: "inherit" }
    );
  } else {
    execFileSync(
      "tar",
      ["-xzf", archivePath, "-C", installDir],
      { stdio: "inherit" }
    );
  }
}

function findBinary(dir) {
  const direct = path.join(dir, binName);
  if (fs.existsSync(direct)) {
    return direct;
  }
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      const found = findBinary(full);
      if (found) {
        return found;
      }
    } else if (entry.name === binName) {
      return full;
    }
  }
  return null;
}

function tryLocalBuild() {
  const goMod = path.join(root, "go.mod");
  if (!fs.existsSync(goMod)) {
    return false;
  }
  const go = spawnSync("go", ["version"], { encoding: "utf8" });
  if (go.status !== 0) {
    return false;
  }
  console.log("Release asset missing; building from local source...");
  fs.mkdirSync(installDir, { recursive: true });
  execFileSync(
    "go",
    ["build", "-o", binPath, "./cmd/skill-man"],
    { stdio: "inherit", cwd: root, env: process.env }
  );
  if (process.platform !== "win32" && fs.existsSync(binPath)) {
    fs.chmodSync(binPath, 0o755);
  }
  return fs.existsSync(binPath);
}

function tryGoInstall() {
  const go = spawnSync("go", ["version"], { encoding: "utf8" });
  if (go.status !== 0) {
    return false;
  }
  const refs = [tag, "latest"];
  for (const ref of refs) {
    console.log(`Release asset missing; trying go install @${ref}...`);
    const result = spawnSync(
      "go",
      ["install", `${MODULE}@${ref}`],
      { stdio: "inherit", env: process.env }
    );
    if (result.status !== 0) {
      continue;
    }
    const gopath = execFileSync("go", ["env", "GOPATH"], {
      encoding: "utf8",
    }).trim();
    const built = path.join(
      gopath,
      "bin",
      process.platform === "win32" ? "skill-man.exe" : "skill-man"
    );
    if (!fs.existsSync(built)) {
      continue;
    }
    fs.mkdirSync(installDir, { recursive: true });
    fs.copyFileSync(built, binPath);
    if (process.platform !== "win32") {
      fs.chmodSync(binPath, 0o755);
    }
    return true;
  }
  return false;
}

async function main() {
  if (fs.existsSync(binPath)) {
    return;
  }

  const { goos, goarch } = platformArch();
  const asset = assetName(version, goos, goarch);
  const url = `https://github.com/${REPO}/releases/download/${tag}/${asset}`;

  try {
    console.log(`Downloading ${asset}...`);
    const data = await download(url);
    const archivePath = path.join(installDir, asset);
    fs.mkdirSync(installDir, { recursive: true });
    fs.writeFileSync(archivePath, data);
    extractArchive(archivePath, goos);
    fs.unlinkSync(archivePath);

    const found = findBinary(installDir);
    if (!found) {
      throw new Error(`binary ${binName} not found in archive`);
    }
    if (found !== binPath) {
      fs.renameSync(found, binPath);
    }
    if (process.platform !== "win32") {
      fs.chmodSync(binPath, 0o755);
    }
    console.log("skill-man installed successfully");
  } catch (err) {
    if (tryLocalBuild()) {
      console.log("skill-man installed via local go build");
      return;
    }
    if (tryGoInstall()) {
      console.log("skill-man installed via go install");
      return;
    }
    console.error(`skill-man install failed: ${err.message}`);
    console.error(
      `\nInstall options:\n  npm install -g skill-man   (needs GitHub release ${tag})\n  go install ${MODULE}@${tag}`
    );
    process.exit(1);
  }
}

main();
