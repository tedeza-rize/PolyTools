import { useCallback, useEffect, useState } from "react";
import {
  Button,
  Card,
  Checkbox,
  Dropdown,
  Input,
  Option,
} from "@fluentui/react-components";
import { AddRegular, DeleteRegular } from "@fluentui/react-icons";
import { PolyToolsService } from "../../../bindings/polytools/internal/services";
import type {
  AutomationRule,
} from "../../../bindings/polytools/internal/modules/models";
import type {
  WinScreensaverInfo,
} from "../../../bindings/polytools/internal/services/models";
import { ModuleInfo } from "../../types";
import { useT } from "../../i18n";

type PanelProps = { module: ModuleInfo };

function msg(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

// --- Automations rules editor ---

const STEP_TYPES = ["run", "open", "type", "keys", "delay", "beep"];
const TRIGGER_TYPES = ["hotkey", "interval", "startup"];

export function AutomationsPanel(_props: PanelProps) {
  const t = useT();
  const [rules, setRules] = useState<AutomationRule[]>([]);
  const [error, setError] = useState("");

  const refresh = useCallback(() => {
    PolyToolsService.AutomationRules()
      .then((r) => setRules(r ?? []))
      .catch((e) => setError(msg(e)));
  }, []);

  useEffect(refresh, [refresh]);

  const persist = async (next: AutomationRule[]) => {
    setRules(next);
    try {
      await PolyToolsService.SaveAutomationRules(next);
    } catch (e) {
      setError(msg(e));
    }
  };

  const addRule = () =>
    persist([
      ...rules,
      {
        id: `rule-${Date.now()}`,
        name: `Rule ${rules.length + 1}`,
        enabled: true,
        trigger: { type: "hotkey", value: "ctrl+alt+1" },
        steps: [{ type: "run", value: "notepad" }],
      },
    ]);

  const update = (id: string, patch: Partial<AutomationRule>) =>
    persist(rules.map((r) => (r.id === id ? { ...r, ...patch } : r)));

  const remove = (id: string) => persist(rules.filter((r) => r.id !== id));

  const stepPlaceholder = (type: string) =>
    ({
      run: "notepad.exe",
      open: "https://example.com",
      type: "text to type",
      keys: "ctrl+v",
      delay: "500",
      beep: "",
    }[type] ?? "");

  const triggerPlaceholder = (type: string) =>
    ({ hotkey: "ctrl+alt+1", interval: "5", startup: "" }[type] ?? "");

  return (
    <>
      <div className="section-label">{t("panel.auto", "Automation rules")}</div>
      <Card className="setting-group" size="small">
        {rules.map((r) => (
          <div className="auto-rule" key={r.id}>
            <div className="auto-rule-head">
              <Checkbox
                checked={r.enabled}
                onChange={(_, d) => update(r.id, { enabled: !!d.checked })}
              />
              <Input
                size="small" value={r.name}
                onChange={(_, d) => update(r.id, { name: d.value })}
                style={{ width: 180 }}
              />
              <Dropdown
                size="small" selectedOptions={[r.trigger.type]}
                value={r.trigger.type}
                onOptionSelect={(_, d) =>
                  update(r.id, { trigger: { type: String(d.optionValue), value: "" } })
                }
                style={{ minWidth: 110 }}
              >
                {TRIGGER_TYPES.map((ty) => (
                  <Option key={ty} value={ty}>{ty}</Option>
                ))}
              </Dropdown>
              {r.trigger.type !== "startup" && (
                <Input
                  size="small" value={r.trigger.value}
                  placeholder={triggerPlaceholder(r.trigger.type)}
                  onChange={(_, d) =>
                    update(r.id, { trigger: { ...r.trigger, value: d.value } })
                  }
                  style={{ width: 130 }}
                />
              )}
              <Button
                size="small" appearance="subtle" icon={<DeleteRegular />}
                onClick={() => remove(r.id)}
              />
            </div>
            <div className="auto-steps">
              {(r.steps ?? []).map((s, i) => (
                <div className="auto-step" key={i}>
                  <Dropdown
                    size="small" selectedOptions={[s.type]} value={s.type}
                    onOptionSelect={(_, d) => {
                      const steps = (r.steps ?? []).slice();
                      steps[i] = { type: String(d.optionValue), value: "" };
                      update(r.id, { steps });
                    }}
                    style={{ minWidth: 90 }}
                  >
                    {STEP_TYPES.map((ty) => (
                      <Option key={ty} value={ty}>{ty}</Option>
                    ))}
                  </Dropdown>
                  {s.type !== "beep" && (
                    <Input
                      size="small" value={s.value}
                      placeholder={stepPlaceholder(s.type)}
                      onChange={(_, d) => {
                        const steps = (r.steps ?? []).slice();
                        steps[i] = { ...s, value: d.value };
                        update(r.id, { steps });
                      }}
                      style={{ flex: 1 }}
                    />
                  )}
                  <Button
                    size="small" appearance="subtle" icon={<DeleteRegular />}
                    onClick={() =>
                      update(r.id, { steps: (r.steps ?? []).filter((_, j) => j !== i) })
                    }
                  />
                </div>
              ))}
              <Button
                size="small" icon={<AddRegular />}
                onClick={() =>
                  update(r.id, {
                    steps: [...(r.steps ?? []), { type: "run", value: "" }],
                  })
                }
              >
                {t("panel.auto.addStep", "Add step")}
              </Button>
            </div>
          </div>
        ))}
        {error && <div className="panel-status panel-error">{error}</div>}
        <div className="panel-actions">
          <Button appearance="primary" size="small" icon={<AddRegular />} onClick={addRule}>
            {t("panel.auto.addRule", "Add rule")}
          </Button>
        </div>
      </Card>
    </>
  );
}

// --- Windows screensaver status ---

export function ScreensaverPanel({ module }: PanelProps) {
  const t = useT();
  const [info, setInfo] = useState<WinScreensaverInfo | null>(null);

  useEffect(() => {
    PolyToolsService.WindowsScreensaver()
      .then(setInfo)
      .catch(console.error);
    // Re-fetch when the module is toggled — takeover suspends/restores
    // the Windows screensaver as a side effect.
  }, [module.enabled]);

  let status = "";
  if (info) {
    if (info.suspended) {
      status = t("panel.ss.suspended", "Suspended by PolyTools");
    } else if (info.active) {
      const name = info.scrPath
        ? info.scrPath.split(/[\\/]/).pop()
        : t("panel.ss.blank", "Blank screen");
      const mins = Math.round(info.timeout / 60);
      status = `${name} — ${t("panel.ss.everyMin", "{n} min").replace("{n}", String(mins))}`;
    } else {
      status = t("panel.ss.off", "Off");
    }
  }

  return (
    <Card className="setting-group" size="small">
      <div className="setting-row">
        <div className="setting-text">
          <div className="setting-label">
            {t("panel.ss.winss", "Windows screensaver")}
          </div>
          <div className="setting-desc">{status || "…"}</div>
        </div>
      </div>
    </Card>
  );
}
