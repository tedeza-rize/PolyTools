import { Badge, Card, Switch, Title2, Text } from "@fluentui/react-components";
import { ModuleInfo } from "../types";
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
  return (
    <div className="content-inner">
      <Title2 className="page-title" as="h1" block>
        Dashboard
      </Title2>
      <Text className="page-subtitle" as="p" block>
        Turn utilities on or off. Click a card to configure it.
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
              <span className="module-card-icon">
                <ModuleIcon name={m.icon} />
              </span>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div className="module-card-name">
                  {m.name}{" "}
                  {!m.available && (
                    <Badge size="small" color="informative">
                      Soon
                    </Badge>
                  )}
                </div>
                <div className="module-card-desc">{m.description}</div>
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
