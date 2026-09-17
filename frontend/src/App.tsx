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
import { General as GeneralSettings, ModuleInfo } from "./types";

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
  const [modules, setModules] = useState<ModuleInfo[]>([]);
  const [general, setGeneral] = useState<GeneralSettings>({
    runAtStartup: false,
    theme: "system",
  });
  const [query, setQuery] = useState("");
  const [maximised, setMaximised] = useState(false);

  const systemDark = useSystemDark();
  const dark =
    general.theme === "dark" ||
    (general.theme === "system" && systemDark);

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
        m.description.toLowerCase().includes(q)
    );
  }, [modules, query]);

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

  return (
    <FluentProvider
      theme={dark ? webDarkTheme : webLightTheme}
      style={{ background: "transparent", height: "100%" }}
    >
      <div className="app-shell">
        <TitleBar
          query={query}
          onQuery={setQuery}
          maximised={maximised}
          onMaximise={() => {
            PolyToolsService.ToggleMaximise().catch(console.error);
            setMaximised((m) => !m);
          }}
        />
        <div className="app-body">
          <SideNav modules={filtered} route={route} onNavigate={setRoute} />
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
    </FluentProvider>
  );
}
