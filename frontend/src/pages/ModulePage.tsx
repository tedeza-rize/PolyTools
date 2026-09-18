import { Badge, Card, Switch } from "@fluentui/react-components";
import { ChevronRightRegular } from "@fluentui/react-icons";
import { CATEGORY_LABELS, ModuleInfo, SettingField } from "../types";
import { useT } from "../i18n";
import { ModuleIcon, moduleColor } from "../components/ModuleIcon";
import { SettingControl } from "../components/SettingControl";
import {
  AutomationsPanel,
  ScreensaverPanel,
} from "../components/panels/Panels";

// Per-module rich content rendered between the hero card and the settings list.
const MODULE_EXTRAS: Record<string, React.ComponentType<{ module: ModuleInfo }>> = {
  "automations": AutomationsPanel,
  "screensaver": ScreensaverPanel,
};

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
  const color = moduleColor(module.icon);

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

  const settings = (module.settings ?? []).map(localize);

  return (
    <div className="content-inner">
      <div className="breadcrumb">
        <span className="breadcrumb-part">
          {t(`cat.${module.category}`, CATEGORY_LABELS[module.category])}
        </span>
        <ChevronRightRegular className="breadcrumb-sep" />
        <span className="breadcrumb-part current">
          {t(`module.${module.key}.name`, module.name)}
        </span>
      </div>

      <Card
        className="module-header"
        style={{ flexDirection: "row", alignItems: "center" }}
      >
        <span className="icon-tile" style={{ ["--tile-color" as any]: color }}>
          <ModuleIcon name={module.icon} />
        </span>
        <div className="module-header-text">
          <div className="module-header-name">
            {t(`module.${module.key}.name`, module.name)}{" "}
            {!module.available && (
              <Badge size="medium" color="informative">
                {t("badge.comingSoon")}
              </Badge>
            )}
          </div>
          <div className="module-header-desc">
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

      {(() => {
        const Extra = MODULE_EXTRAS[module.key];
        return Extra ? <Extra module={module} /> : null;
      })()}

      {settings.length > 0 && (
        <>
          <div className="section-label">{t("module.settings")}</div>
          <Card className="setting-group" size="small">
            {settings.map((f) => (
              <div className="setting-row" key={f.key}>
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
              </div>
            ))}
          </Card>
        </>
      )}
    </div>
  );
}
