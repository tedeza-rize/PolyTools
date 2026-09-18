import { useCallback, useEffect, useMemo, useState } from "react";
import {
  FluentProvider,
  webDarkTheme,
  webLightTheme,
} from "@fluentui/react-components";
import { Events } from "@wailsio/runtime";
import { PolyToolsService } from "../bindings/polytools/internal/services";
import { TitleBar } from "./components/TitleBar";
import { SideNav, Route } from "./components/SideNav";
import { Dashboard } from "./pages/Dashboard";
import { General } from "./pages/General";
import { ModulePage } from "./pages/ModulePage";
import { ScreensaverPage } from "./pages/Tools";
import { General as GeneralSettings, ModuleInfo } from "./types";
import { I18nProvider, translatorFor } from "./i18n";

// Dedicated overlay windows (screensaver) load the app with ?page=... —
// render those without the settings shell.
const TOOL_PAGES: Record<string, React.ComponentType> = {
  screensaver: ScreensaverPage,
};

function toolPage(): React.ComponentType | null {
  const page = new URLSearchParams(window.location.search).get("page");
  return page ? TOOL_PAGES[page] ?? null : null;
}

function useSystemDark() {
  const [dark, setDark] = useState(
    () => window.matchMedia("(prefers-color-scheme: dark)").matches
  );
  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e: MediaQueryListEvent) => setDark(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);
  return dark;
}

export default function App() {
  const [route, setRoute] = useState<Route>("dashboard");
  const tool = useMemo(toolPage, []);
  const [modules, setModules] = useState<ModuleInfo[]>([]);
  const [general, setGeneral] = useState<GeneralSettings>({
    runAtStartup: false,
    theme: "system",
    language: "system",
  });
  const [query, setQuery] = useState("");
  const [maximised, setMaximised] = useState(false);

  const systemDark = useSystemDark();
  const dark =
    general.theme === "dark" ||
    (general.theme === "system" && systemDark);

  const t = useMemo(() => translatorFor(general.language), [general.language]);

  const refresh = useCallback(() => {
    PolyToolsService.Modules()
      .then((m) => setModules(m ?? []))
      .catch(console.error);
  }, []);

  useEffect(() => {
    refresh();
    PolyToolsService.General().then(setGeneral).catch(console.error);
    return Events.On("modules:changed", refresh);
  }, [refresh]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return modules;
    return modules.filter(
      (m) =>
        m.name.toLowerCase().includes(q) ||
        m.description.toLowerCase().includes(q) ||
        t(`module.${m.key}.name`, m.name).toLowerCase().includes(q) ||
        t(`module.${m.key}.desc`, m.description).toLowerCase().includes(q)
    );
  }, [modules, query, t]);

  const toggle = useCallback((key: string, enabled: boolean) => {
    PolyToolsService.SetModuleEnabled(key, enabled)
      .then(refresh)
      .catch(console.error);
  }, [refresh]);

  const updateSetting = useCallback(
    (key: string, field: string, value: any) => {
      PolyToolsService.UpdateModuleSetting(key, field, value)
        .then(refresh)
        .catch(console.error);
    },
    [refresh]
  );

  const openModule = useCallback((key: string) => {
    setRoute(`module:${key}`);
  }, []);

  const currentModule = route.startsWith("module:")
    ? modules.find((m) => m.key === route.slice(7))
    : undefined;

  if (tool) {
    const Tool = tool;
    return (
      <FluentProvider
        theme={dark ? webDarkTheme : webLightTheme}
        style={{ background: "transparent", height: "100%" }}
      >
        <Tool />
      </FluentProvider>
    );
  }

  return (
    <FluentProvider
      theme={dark ? webDarkTheme : webLightTheme}
      style={{ background: "transparent", height: "100%" }}
    >
      <I18nProvider value={t}>
      <div className="app-shell">
        <TitleBar
          canGoBack={route !== "dashboard"}
          onBack={() => setRoute("dashboard")}
          maximised={maximised}
          onMaximise={() => {
            PolyToolsService.ToggleMaximise().catch(console.error);
            setMaximised((m) => !m);
          }}
        />
        <div className="app-body">
          <SideNav
            modules={filtered}
            route={route}
            query={query}
            onQuery={setQuery}
            onNavigate={setRoute}
          />
          <main className="content">
            {route === "dashboard" && (
              <Dashboard
                modules={filtered}
                onToggle={toggle}
                onOpen={openModule}
              />
            )}
            {route === "general" && (
              <General
                general={general}
                onRunAtStartup={(v) => {
                  PolyToolsService.SetRunAtStartup(v)
                    .then(() =>
                      setGeneral((g) => ({ ...g, runAtStartup: v }))
                    )
                    .catch(console.error);
                }}
                onTheme={(v) => {
                  PolyToolsService.SetTheme(v)
                    .then(() => setGeneral((g) => ({ ...g, theme: v })))
                    .catch(console.error);
                }}
                onLanguage={(v) => {
                  PolyToolsService.SetLanguage(v)
                    .then(() => setGeneral((g) => ({ ...g, language: v })))
                    .catch(console.error);
                }}
              />
            )}
            {currentModule && (
              <ModulePage
                module={currentModule}
                onToggle={toggle}
                onSetting={updateSetting}
              />
            )}
          </main>
        </div>
      </div>
      </I18nProvider>
    </FluentProvider>
  );
}
