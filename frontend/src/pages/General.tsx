import {
  Card,
  Dropdown,
  Option,
  Switch,
  Text,
  Title2,
} from "@fluentui/react-components";
import { General as GeneralSettings } from "../types";
import { useT } from "../i18n";

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

export function General({
  general,
  onRunAtStartup,
  onTheme,
  onLanguage,
}: {
  general: GeneralSettings;
  onRunAtStartup: (v: boolean) => void;
  onTheme: (v: string) => void;
  onLanguage: (v: string) => void;
}) {
  const t = useT();

  const THEMES = [
    { value: "system", label: t("general.theme.system") },
    { value: "light", label: t("general.theme.light") },
    { value: "dark", label: t("general.theme.dark") },
  ];

  const LANGS = [
    { value: "system", label: t("general.lang.system") },
    { value: "en", label: "English" },
    { value: "ko", label: "한국어" },
  ];

  return (
    <div className="content-inner">
      <Title2 className="page-title" as="h1" block>
        {t("general.title")}
      </Title2>
      <Text className="page-subtitle" as="p" block>
        {t("general.subtitle")}
      </Text>

      <div className="section-label">{t("general.startup")}</div>
      <Row
        label={t("general.runAtStartup")}
        desc={t("general.runAtStartup.desc")}
        control={
          <Switch
            checked={general.runAtStartup}
            onChange={(_, d) => onRunAtStartup(d.checked)}
          />
        }
      />

      <div className="section-label">{t("general.appearance")}</div>
      <Row
        label={t("general.theme")}
        desc={t("general.theme.desc")}
        control={
          <Dropdown
            size="small"
            selectedOptions={[general.theme]}
            value={THEMES.find((x) => x.value === general.theme)?.label}
            onOptionSelect={(_, d) => d.optionValue && onTheme(d.optionValue)}
            style={{ minWidth: 180 }}
          >
            {THEMES.map((x) => (
              <Option key={x.value} value={x.value}>
                {x.label}
              </Option>
            ))}
          </Dropdown>
        }
      />
      <Row
        label={t("general.language")}
        desc={t("general.language.desc")}
        control={
          <Dropdown
            size="small"
            selectedOptions={[general.language]}
            value={LANGS.find((x) => x.value === general.language)?.label}
            onOptionSelect={(_, d) => d.optionValue && onLanguage(d.optionValue)}
            style={{ minWidth: 180 }}
          >
            {LANGS.map((x) => (
              <Option key={x.value} value={x.value}>
                {x.label}
              </Option>
            ))}
          </Dropdown>
        }
      />

      <div className="section-label">{t("general.about")}</div>
      <Card
        size="small"
        className="setting-card"
        style={{ flexDirection: "row", alignItems: "center", gap: 16 }}
      >
        <div className="setting-text">
          <div className="setting-label">PolyTools</div>
          <div className="setting-desc">{t("general.about.desc")}</div>
        </div>
      </Card>
    </div>
  );
}
