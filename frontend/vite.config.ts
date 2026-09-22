import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
      "@bindings": path.resolve(import.meta.dirname, "./bindings"),
      "@locales": path.resolve(import.meta.dirname, "./src/locales"),
    },
  },
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [
    tailwindcss(),
    react(),
    wails("./bindings"),
  ],
  build: {
    chunkSizeWarningLimit: 1500,
  },
});