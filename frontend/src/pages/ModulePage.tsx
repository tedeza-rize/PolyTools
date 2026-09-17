import { Badge, Card, Switch } from "@fluentui/react-components";
import { ModuleInfo } from "../types";
import { ModuleIcon } from "../components/ModuleIcon";
import { SettingControl } from "../components/SettingControl";

export function ModulePage({
  module,
  onToggle,
  onSetting,
}: {
  module: ModuleInfo;
  onToggle: (key: string, enabled: boolean) => void;
  onSetting: (key: string, field: string, value: any) => void;
}) {
  const disabled = !module.available || !module.enabled;

  return (
    <div className="content-inner">
      <Card
        className="module-hero"
        style={{ flexDirection: "row", alignItems: "center", gap: 16 }}
      >
        <span className="module-hero-icon">
          <ModuleIcon name={module.icon} />
        </span>
        <div className="module-hero-text">
          <div className="module-hero-name">
            {module.name}{" "}
            {!module.available && (
              <Badge size="medium" color="informative">
                Coming soon
              </Badge>
            )}
          </div>
          <div className="module-hero-desc">{module.description}</div>
        </div>
        <Switch
          checked={module.enabled}
          disabled={!module.available}
          onChange={(_, d) => onToggle(module.key, d.checked)}
          label={module.enabled ? "On" : "Off"}
        />
      </Card>

      {module.settings && module.settings.length > 0 && (
        <>
          <div className="section-label">Settings</div>
          {module.settings.map((f) => (
            <Card
              size="small"
              className="setting-card"
              key={f.key}
              style={{ flexDirection: "row", alignItems: "center", gap: 16 }}
            >
              <div className="setting-text">
                <div className="setting-label">{f.label}</div>
                {f.description && (
                  <div className="setting-desc">{f.description}</div>
                )}
              </div>
              <SettingControl
                field={f}
                disabled={disabled}
                onChange={(v) => onSetting(module.key, f.key, v)}
              />
            </Card>
          ))}
        </>
      )}
    </div>
  );
}
