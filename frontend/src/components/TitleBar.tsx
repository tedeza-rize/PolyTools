import { SearchBox } from "@fluentui/react-components";
import {
  AppsListDetailRegular,
  SubtractRegular,
  SquareRegular,
  SquareMultipleRegular,
  DismissRegular,
} from "@fluentui/react-icons";
import { PolyToolsService } from "../../bindings/polytools/internal/services";

export function TitleBar({
  query,
  onQuery,
  maximised,
  onMaximise,
}: {
  query: string;
  onQuery: (q: string) => void;
  maximised: boolean;
  onMaximise: () => void;
}) {
  return (
    <div className="titlebar">
      <div className="titlebar-brand">
        <span className="titlebar-icon">
          <AppsListDetailRegular />
        </span>
        <span>PolyTools</span>
      </div>

      <div className="titlebar-search">
        <div className="titlebar-search-inner">
          <SearchBox
            size="small"
            placeholder="Find a utility or setting"
            value={query}
            onChange={(_, d) => onQuery(d.value)}
            style={{ width: "100%" }}
          />
        </div>
      </div>

      <div className="caption-buttons">
        <button
          className="caption-btn"
          style={{ ["--wails-non-client-region" as any]: "minimize" }}
          onClick={() => PolyToolsService.Minimise()}
          aria-label="Minimize"
        >
          <SubtractRegular />
        </button>
        <button
          className="caption-btn"
          style={{ ["--wails-non-client-region" as any]: "maximize" }}
          onClick={onMaximise}
          aria-label={maximised ? "Restore" : "Maximize"}
        >
          {maximised ? <SquareMultipleRegular /> : <SquareRegular />}
        </button>
        <button
          className="caption-btn close"
          style={{ ["--wails-non-client-region" as any]: "close" }}
          onClick={() => PolyToolsService.HideWindow()}
          aria-label="Close"
        >
          <DismissRegular />
        </button>
      </div>
    </div>
  );
}
