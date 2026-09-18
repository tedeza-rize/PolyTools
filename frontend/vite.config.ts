import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import wails from "@wailsio/runtime/plugins/vite";
import http from "node:http";
import net from "node:net";

const PORT = Number(process.env.WAILS_VITE_PORT) || 9245;

// Wails dev points webviews at http://localhost:<port>. `localhost` may
// resolve to ::1 or 127.0.0.1 depending on the client, and a refused first
// attempt can flash an error page in the webview. Vite binds ::1; this
// plugin answers on 127.0.0.1 too and forwards traffic (incl. HMR
// websocket upgrades) to ::1 so both loopbacks always work.
function dualLoopback(port: number): Plugin {
  return {
    name: "dual-loopback",
    configureServer() {
      const srv = http.createServer((req, res) => {
        const up = http.request(
          { host: "::1", port, path: req.url, method: req.method, headers: req.headers },
          (r) => {
            res.writeHead(r.statusCode ?? 502, r.headers);
            r.pipe(res);
          },
        );
        up.on("error", () => {
          res.writeHead(502);
          res.end();
        });
        req.pipe(up);
      });
      srv.on("upgrade", (req, socket, head) => {
        const up = net.connect(port, "::1", () => {
          const headers = Object.entries(req.headers)
            .map(([k, v]) => `${k}: ${v}`)
            .join("\r\n");
          up.write(`${req.method} ${req.url} HTTP/1.1\r\n${headers}\r\n\r\n`);
          if (head?.length) up.write(head);
          up.pipe(socket).pipe(up);
        });
        up.on("error", () => socket.destroy());
      });
      srv.listen(port, "127.0.0.1");
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "localhost",
    port: PORT,
    strictPort: true,
  },
  plugins: [react(), wails("./bindings"), dualLoopback(PORT)],
});
