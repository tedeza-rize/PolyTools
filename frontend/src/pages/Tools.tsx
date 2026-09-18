import { useEffect, useRef, useState } from "react";
import { PolyToolsService } from "../../bindings/polytools/internal/services";

// Fullscreen tool page rendered in a dedicated overlay window
// (opened by the backend via /?page=screensaver).

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
