import { Badge, Card, Switch } from "@fluentui/react-components";
import { ModuleInfo, SettingField } from "../types";
import { useT } from "../i18n";
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
  const t = useT();
  const disabled = !module.available || !module.enabled;

  // Translate declarative field labels/descriptions/option labels with
  // fallback to the backend-provided English text.
  const localize = (f: SettingField): SettingField => ({
    ...f,
    label: t(`set.${module.key}.${f.key}.label`, f.label),
    description: f.description
      ? t(`set.${module.key}.${f.key}.desc`, f.description)
      : f.description,
    options: f.options?.map((o) => ({
      ...o,
      label: t(`set.${module.key}.${f.key}.opt.${o.value}`, o.label),
    })),
  });

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
            {t(`module.${module.key}.name`, module.name)}{" "}
            {!module.available && (
              <Badge size="medium" color="informative">
                {t("badge.comingSoon")}
              </Badge>
            )}
          </div>
          <div className="module-hero-desc">
            {t(`module.${module.key}.desc`, module.description)}
          </div>
        </div>
        <Switch
          checked={module.enabled}
          disabled={!module.available}
          onChange={(_, d) => onToggle(module.key, d.checked)}
          label={module.enabled ? t("module.on") : t("module.off")}
        />
      </Card>

      {module.settings && module.settings.length > 0 && (
        <>
          <div className="section-label">{t("module.settings")}</div>
          {module.settings.map(localize).map((f) => (
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
