import { useCallback, useEffect, useRef, useState } from "react";
import { FluentProvider, webDarkTheme } from "@fluentui/react-components";
import { PolyToolsService } from "../../bindings/polytools/internal/services";

// Fullscreen tool pages rendered in dedicated overlay windows
// (opened by backend hotkeys via /?page=ruler|extract|peek|screensaver).

function useCloseOnEscape() {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") PolyToolsService.CloseToolWindow().catch(() => {});
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);
}

// --- Screen Ruler: drag to measure px distance ---

export function RulerPage() {
  const [start, setStart] = useState<{ x: number; y: number } | null>(null);
  const [cur, setCur] = useState<{ x: number; y: number } | null>(null);
  useCloseOnEscape();

  const dist =
    start && cur
      ? Math.round(Math.hypot(cur.x - start.x, cur.y - start.y))
      : null;
  const dx = start && cur ? Math.abs(cur.x - start.x) : 0;
  const dy = start && cur ? Math.abs(cur.y - start.y) : 0;

  return (
    <div
      className="tool-overlay"
      onMouseDown={(e) => {
        setStart({ x: e.clientX, y: e.clientY });
        setCur({ x: e.clientX, y: e.clientY });
      }}
      onMouseMove={(e) => start && setCur({ x: e.clientX, y: e.clientY })}
      onMouseUp={() => setStart(null)}
    >
      <div className="tool-hint">드래그하여 측정 · ESC로 닫기</div>
      {cur && <div className="tool-crosshair-v" style={{ left: cur.x }} />}
      {cur && <div className="tool-crosshair-h" style={{ top: cur.y }} />}
      {start && cur && (
        <>
          <svg className="tool-svg">
            <line
              x1={start.x}
              y1={start.y}
              x2={cur.x}
              y2={cur.y}
              stroke="#ffd400"
              strokeWidth="2"
            />
          </svg>
          <div
            className="tool-measure"
            style={{
              left: (start.x + cur.x) / 2,
              top: Math.min(start.y, cur.y) - 44,
            }}
          >
            {dist}px
            <span className="tool-measure-sub">
              {Math.round(dx)} × {Math.round(dy)}
            </span>
          </div>
        </>
      )}
    </div>
  );
}

// --- Text Extractor: drag a region → OCR → clipboard ---

export function ExtractPage() {
  const [start, setStart] = useState<{ x: number; y: number } | null>(null);
  const [cur, setCur] = useState<{ x: number; y: number } | null>(null);
  const [busy, setBusy] = useState(false);
  useCloseOnEscape();

  const finish = useCallback(
    async (e: React.MouseEvent) => {
      if (!start) return;
      const x = Math.min(start.x, e.clientX);
      const y = Math.min(start.y, e.clientY);
      const w = Math.abs(e.clientX - start.x);
      const h = Math.abs(e.clientY - start.y);
      setStart(null);
      setCur(null);
      if (w < 8 || h < 8) return;
      setBusy(true);
      try {
        // Window origin in screen coords; CSS px == physical px (PerMonitorV2).
        const originX = e.screenX - e.clientX;
        const originY = e.screenY - e.clientY;
        await PolyToolsService.ExtractRegion(
          Math.round(originX + x),
          Math.round(originY + y),
          Math.round(w),
          Math.round(h)
        );
      } catch (err) {
        console.error(err);
      } finally {
        setBusy(false);
        PolyToolsService.CloseToolWindow().catch(() => {});
      }
    },
    [start]
  );

  const rect =
    start && cur
      ? {
          left: Math.min(start.x, cur.x),
          top: Math.min(start.y, cur.y),
          width: Math.abs(cur.x - start.x),
          height: Math.abs(cur.y - start.y),
        }
      : null;

  return (
    <div
      className="tool-overlay"
      onMouseDown={(e) => {
        setStart({ x: e.clientX, y: e.clientY });
        setCur({ x: e.clientX, y: e.clientY });
      }}
      onMouseMove={(e) => start && setCur({ x: e.clientX, y: e.clientY })}
      onMouseUp={finish}
    >
      <div className="tool-hint">
        {busy ? "인식 중…" : "영역을 드래그하여 텍스트 추출 · ESC로 취소"}
      </div>
      {rect && <div className="tool-select-rect" style={rect} />}
    </div>
  );
}

// --- Quick Peek: preview a file path or clipboard ---

type PeekResult = {
  path: string;
  kind: string;
  name: string;
  text: string;
  size: number;
  modified: string;
};

export function PeekPage() {
  const [result, setResult] = useState<PeekResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const path = params.get("path") ?? "";
    PolyToolsService.PeekFile(path)
      .then((r) => setResult(r as PeekResult))
      .catch((e) => setError(String(e)));
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") window.close();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <FluentProvider theme={webDarkTheme}>
      <div className="peek-page">
        <div className="peek-title">
          {result?.name || "Quick Peek"}
          {result ? (
            <span className="peek-meta">
              {result.size.toLocaleString()} bytes
            </span>
          ) : null}
        </div>
        <div className="peek-body">
          {error && <div className="peek-text">{error}</div>}
          {result?.kind === "image" && (
            <img className="peek-image" src={result.text} alt={result.name} />
          )}
          {(result?.kind === "text" ||
            result?.kind === "binary" ||
            result?.kind === "clipboard") && (
            <pre className="peek-text">
              {result.text || "(클립보드가 비어 있습니다)"}
            </pre>
          )}
          {!result && !error && <div className="peek-text">불러오는 중…</div>}
        </div>
      </div>
    </FluentProvider>
  );
}

// --- Screensaver: fullscreen content ---

export function ScreensaverPage() {
  const [urls, setUrls] = useState<string[]>([]);
  const [type, setType] = useState("video");
  const [idx, setIdx] = useState(0);
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const t = params.get("type") ?? "video";
    const src = params.get("src") ?? "";
    setType(t);
    PolyToolsService.ScreensaverMedia(t, src)
      .then((u) => setUrls(u ?? []))
      .catch(console.error);
    const dismiss = () =>
      PolyToolsService.DismissScreensaver().catch(() => {});
    window.addEventListener("keydown", dismiss);
    window.addEventListener("mousedown", dismiss);
    return () => {
      window.removeEventListener("keydown", dismiss);
      window.removeEventListener("mousedown", dismiss);
    };
  }, []);

  useEffect(() => {
    if (type !== "images" || urls.length < 2) return;
    const iv = setInterval(() => setIdx((i) => (i + 1) % urls.length), 8000);
    return () => clearInterval(iv);
  }, [type, urls.length]);

  const src = urls.length ? urls[idx % urls.length] : null;

  return (
    <div className="screensaver-page">
      {!src && <div className="screensaver-fallback">PolyTools</div>}
      {src && type === "video" && (
        <video
          ref={videoRef}
          className="screensaver-media"
          src={src}
          autoPlay
          loop
          muted
        />
      )}
      {src && type === "web" && (
        <iframe className="screensaver-media" src={src} title="screensaver" />
      )}
      {src && type === "images" && (
        <img className="screensaver-media" src={src} alt="" />
      )}
    </div>
  );
}
