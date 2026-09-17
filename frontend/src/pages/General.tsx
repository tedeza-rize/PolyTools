import {
  Card,
  Dropdown,
  Option,
  Switch,
  Text,
  Title2,
} from "@fluentui/react-components";
import { General as GeneralSettings } from "../types";

function Row({
  label,
  desc,
  control,
}: {
  label: string;
  desc?: string;
  control: JSX.Element;
}) {
  return (
    <Card
      size="small"
      className="setting-card"
      style={{ flexDirection: "row", alignItems: "center", gap: 16 }}
    >
      <div className="setting-text">
        <div className="setting-label">{label}</div>
        {desc && <div className="setting-desc">{desc}</div>}
      </div>
      {control}
    </Card>
  );
}

const THEMES = [
  { value: "system", label: "System default" },
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
];

export function General({
  general,
  onRunAtStartup,
  onTheme,
}: {
  general: GeneralSettings;
  onRunAtStartup: (v: boolean) => void;
  onTheme: (v: string) => void;
}) {
  return (
    <div className="content-inner">
      <Title2 className="page-title" as="h1" block>
        General
      </Title2>
      <Text className="page-subtitle" as="p" block>
        App-wide settings.
      </Text>

      <div className="section-label">Startup</div>
      <Row
        label="Run at startup"
        desc="Start PolyTools automatically when you sign in to Windows."
        control={
          <Switch
            checked={general.runAtStartup}
            onChange={(_, d) => onRunAtStartup(d.checked)}
          />
        }
      />

      <div className="section-label">Appearance</div>
      <Row
        label="Theme"
        desc="Choose the app theme or follow Windows."
        control={
          <Dropdown
            size="small"
            selectedOptions={[general.theme]}
            value={THEMES.find((t) => t.value === general.theme)?.label}
            onOptionSelect={(_, d) => d.optionValue && onTheme(d.optionValue)}
            style={{ minWidth: 180 }}
          >
            {THEMES.map((t) => (
              <Option key={t.value} value={t.value}>
                {t.label}
              </Option>
            ))}
          </Dropdown>
        }
      />

      <div className="section-label">About</div>
      <Card
        size="small"
        className="setting-card"
        style={{ flexDirection: "row", alignItems: "center", gap: 16 }}
      >
        <div className="setting-text">
          <div className="setting-label">PolyTools</div>
          <div className="setting-desc">
            Version 0.1.0 — Windows utilities for power users. Built with Go +
            Wails v3.
          </div>
        </div>
      </Card>
    </div>
  );
}
