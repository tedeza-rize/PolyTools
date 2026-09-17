import { Badge, Card, Switch, Title2, Text } from "@fluentui/react-components";
import { ModuleInfo } from "../types";
import { useT } from "../i18n";
import { ModuleIcon } from "../components/ModuleIcon";

export function Dashboard({
  modules,
  onToggle,
  onOpen,
}: {
  modules: ModuleInfo[];
  onToggle: (key: string, enabled: boolean) => void;
  onOpen: (key: string) => void;
}) {
  const t = useT();
  return (
    <div className="content-inner">
      <Title2 className="page-title" as="h1" block>
        {t("dash.title")}
      </Title2>
      <Text className="page-subtitle" as="p" block>
        {t("dash.subtitle")}
      </Text>

      <div className="module-grid">
        {modules.map((m) => (
          <Card
            key={m.key}
            className="module-card"
            size="small"
            onClick={() => onOpen(m.key)}
          >
            <div className="module-card-head">
              <span className="icon-tile">
                <ModuleIcon name={m.icon} />
              </span>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div className="module-card-name">
                  {t(`module.${m.key}.name`, m.name)}{" "}
                  {!m.available && (
                    <Badge size="small" color="informative">
                      {t("badge.soon")}
                    </Badge>
                  )}
                </div>
                <div className="module-card-desc">
                  {t(`module.${m.key}.desc`, m.description)}
                </div>
              </div>
              <Switch
                checked={m.enabled}
                disabled={!m.available}
                onClick={(e) => e.stopPropagation()}
                onChange={(_, d) => onToggle(m.key, d.checked)}
              />
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
