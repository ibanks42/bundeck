import tailwindcss from "@tailwindcss/vite";
import { TanStackRouterVite } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

const ReactCompilerConfig = { target: "19" };

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    TanStackRouterVite({}),
    react({
      babel: {
        plugins: [["babel-plugin-react-compiler", ReactCompilerConfig]],
      },
    }),
    tailwindcss(),
    viteSingleFile(),
  ],
  server: {
    allowedHosts: ["server.shrew-mora.ts.net"],
  },
  preview: {
    allowedHosts: ["server.shrew-mora.ts.net"],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    outDir: "dist",
  },
});
