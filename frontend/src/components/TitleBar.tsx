import {
  AppsListDetailRegular,
  ArrowLeftRegular,
  SubtractRegular,
  SquareRegular,
  SquareMultipleRegular,
  DismissRegular,
} from "@fluentui/react-icons";
import { PolyToolsService } from "../../bindings/polytools/internal/services";
import { useT } from "../i18n";

export function TitleBar({
  canGoBack,
  onBack,
  maximised,
  onMaximise,
}: {
  canGoBack: boolean;
  onBack: () => void;
  maximised: boolean;
  onMaximise: () => void;
}) {
  const t = useT();
  return (
    <div className="titlebar">
      <div className="titlebar-brand">
        {canGoBack && (
          <button
            className="titlebar-back"
            onClick={onBack}
            aria-label={t("nav.back")}
          >
            <ArrowLeftRegular />
          </button>
        )}
        <span className="titlebar-icon">
          <AppsListDetailRegular />
        </span>
        <span>PolyTools</span>
      </div>

      <div className="titlebar-spacer" />

      <div className="caption-buttons">
        <button
          className="caption-btn"
          style={{ ["--wails-non-client-region" as any]: "minimize" }}
          onClick={() => PolyToolsService.Minimise()}
          aria-label={t("caption.minimize")}
        >
          <SubtractRegular />
        </button>
        <button
          className="caption-btn"
          style={{ ["--wails-non-client-region" as any]: "maximize" }}
          onClick={onMaximise}
          aria-label={maximised ? t("caption.restore") : t("caption.maximize")}
        >
          {maximised ? <SquareMultipleRegular /> : <SquareRegular />}
        </button>
        <button
          className="caption-btn close"
          style={{ ["--wails-non-client-region" as any]: "close" }}
          onClick={() => PolyToolsService.HideWindow()}
          aria-label={t("caption.close")}
        >
          <DismissRegular />
        </button>
      </div>
    </div>
  );
}
