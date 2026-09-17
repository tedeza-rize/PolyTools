import { Badge } from "@fluentui/react-components";
import {
  HomeRegular,
  SettingsRegular,
} from "@fluentui/react-icons";
import { CATEGORY_LABELS, CATEGORY_ORDER, ModuleInfo } from "../types";
import { useT } from "../i18n";
import { ModuleIcon } from "./ModuleIcon";

export type Route = "dashboard" | "general" | `module:${string}`;

function NavItem({
  icon,
  label,
  badge,
  selected,
  onClick,
}: {
  icon: JSX.Element;
  label: string;
  badge?: string;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      className={`nav-item${selected ? " selected" : ""}`}
      onClick={onClick}
    >
      <span className="nav-item-icon">{icon}</span>
      <span className="nav-item-label">{label}</span>
      {badge && (
        <Badge size="small" appearance="filled" color="informative">
          {badge}
        </Badge>
      )}
    </button>
  );
}

export function SideNav({
  modules,
  route,
  onNavigate,
}: {
  modules: ModuleInfo[];
  route: Route;
  onNavigate: (r: Route) => void;
}) {
  const t = useT();
  return (
    <nav className="sidenav">
      <NavItem
        icon={<HomeRegular />}
        label={t("nav.dashboard")}
        selected={route === "dashboard"}
        onClick={() => onNavigate("dashboard")}
      />
      <NavItem
        icon={<SettingsRegular />}
        label={t("nav.general")}
        selected={route === "general"}
        onClick={() => onNavigate("general")}
      />

      {CATEGORY_ORDER.map((cat) => {
        const items = modules.filter((m) => m.category === cat);
        if (!items.length) return null;
        return (
          <div className="sidenav-group" key={cat}>
            <div className="sidenav-group-label">
              {t(`cat.${cat}`, CATEGORY_LABELS[cat])}
            </div>
            {items.map((m) => (
              <NavItem
                key={m.key}
                icon={<ModuleIcon name={m.icon} />}
                label={t(`module.${m.key}.name`, m.name)}
                badge={m.available ? undefined : t("badge.soon")}
                selected={route === `module:${m.key}`}
                onClick={() => onNavigate(`module:${m.key}`)}
              />
            ))}
          </div>
        );
      })}
    </nav>
  );
}
