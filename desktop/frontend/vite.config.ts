import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "node:path";

// Wails embeds this build's output (frontend/dist) directly into the Go
// binary via go:embed — base './' keeps asset URLs relative so the same
// build also works served from `vite preview` / a plain file server, which
// is how this UI gets visually QA'd in environments without a display.
export default defineConfig({
  base: "./",
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    outDir: "dist",
  },
});
