import { useCallback, useEffect, useState } from "react";
import {
  Button,
  Card,
  Checkbox,
  Dropdown,
  Input,
  Option,
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableHeaderCell,
  TableRow,
  Textarea,
} from "@fluentui/react-components";
import { AddRegular, DeleteRegular, FolderOpenRegular } from "@fluentui/react-icons";
import { PolyToolsService } from "../../../bindings/polytools/internal/services";
import type {
  AutomationRule,
} from "../../../bindings/polytools/internal/modules/models";
import { ModuleInfo } from "../../types";
import { useT } from "../../i18n";

type PanelProps = { module: ModuleInfo };

function msg(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

// --- Batch Rename ---

type RenameOptions = {
  dir: string;
  pattern: string;
  replace: string;
  useRegex: boolean;
  filterExt: string;
  counter: boolean;
};

export function BatchRenamePanel(_props: PanelProps) {
  const t = useT();
  const [opts, setOpts] = useState<RenameOptions>({
    dir: "", pattern: "", replace: "", useRegex: false, filterExt: "", counter: false,
  });
  const [preview, setPreview] = useState<{ old: string; new: string }[]>([]);
  const [status, setStatus] = useState("");

  const doPreview = useCallback(async () => {
    if (!opts.dir) return;
    try {
      const r = await PolyToolsService.PreviewRename(opts);
      setPreview(r ?? []);
      setStatus(r?.length ? `${r.length}개 파일 변경 예정` : "변경될 파일 없음");
    } catch (e) {
      setStatus(msg(e));
    }
  }, [opts]);

  useEffect(() => {
    if (opts.dir && opts.pattern) doPreview();
    else setPreview([]);
  }, [opts, doPreview]);

  const apply = async () => {
    try {
      const n = await PolyToolsService.ApplyRename(opts);
      setStatus(`${n}개 파일 이름 변경 완료`);
      doPreview();
    } catch (e) {
      setStatus(msg(e));
    }
  };

  return (
    <>
      <div className="section-label">{t("panel.rename", "Batch rename")}</div>
      <Card className="setting-group" size="small">
        <div className="setting-row">
          <div className="setting-text">
            <div className="setting-label">{t("panel.rename.dir", "Folder")}</div>
          </div>
          <div className="panel-row">
            <Input
              size="small" value={opts.dir} placeholder="C:\\…"
              onChange={(_, d) => setOpts({ ...opts, dir: d.value })}
              style={{ minWidth: 260 }}
            />
            <Button
              size="small" icon={<FolderOpenRegular />}
              onClick={() =>
                PolyToolsService.PickFolder("Select folder")
                  .then((p) => p && setOpts((o) => ({ ...o, dir: p })))
                  .catch(console.error)
              }
            />
          </div>
        </div>
        <div className="setting-row">
          <div className="setting-text">
            <div className="setting-label">{t("panel.rename.pattern", "Pattern → Replace")}</div>
            <div className="setting-desc">{t("panel.rename.patternDesc", "Use {n} for a counter. Regex optional.")}</div>
          </div>
          <div className="panel-row">
            <Input
              size="small" value={opts.pattern} placeholder={t("panel.rename.find", "Find")}
              onChange={(_, d) => setOpts({ ...opts, pattern: d.value })}
            />
            <Input
              size="small" value={opts.replace} placeholder={t("panel.rename.replace", "Replace")}
              onChange={(_, d) => setOpts({ ...opts, replace: d.value })}
            />
          </div>
        </div>
        <div className="setting-row">
          <div className="setting-text">
            <div className="setting-label">{t("panel.rename.options", "Options")}</div>
          </div>
          <div className="panel-row">
            <Checkbox
              label="Regex" checked={opts.useRegex}
              onChange={(_, d) => setOpts({ ...opts, useRegex: !!d.checked })}
            />
            <Checkbox
              label="{n}" checked={opts.counter}
              onChange={(_, d) => setOpts({ ...opts, counter: !!d.checked })}
            />
            <Input
              size="small" value={opts.filterExt} placeholder={t("panel.rename.ext", "Ext filter (e.g. .jpg)")}
              onChange={(_, d) => setOpts({ ...opts, filterExt: d.value })}
              style={{ width: 140 }}
            />
          </div>
        </div>
        {preview.length > 0 && (
          <div className="panel-preview">
            {preview.slice(0, 12).map((p) => (
              <div className="panel-preview-row" key={p.old}>
                <span className="panel-old">{p.old}</span>
                <span className="panel-arrow">→</span>
                <span className="panel-new">{p.new}</span>
              </div>
            ))}
            {preview.length > 12 && (
              <div className="panel-preview-row">… +{preview.length - 12}</div>
            )}
          </div>
        )}
        <div className="panel-actions">
          {status && <span className="panel-status">{status}</span>}
          <Button
            appearance="primary" size="small"
            disabled={!preview.length} onClick={apply}
          >
            {t("panel.rename.apply", "Rename files")}
          </Button>
        </div>
      </Card>
    </>
  );
}

// --- Environment Variables ---

type EnvVar = { name: string; value: string; scope: string };

export function EnvVarsPanel(_props: PanelProps) {
  const t = useT();
  const [vars, setVars] = useState<EnvVar[]>([]);
  const [name, setName] = useState("");
  const [value, setValue] = useState("");
  const [scope, setScope] = useState("user");
  const [error, setError] = useState("");

  const refresh = useCallback(() => {
    PolyToolsService.EnvVars()
      .then((v) => setVars(v ?? []))
      .catch((e) => setError(msg(e)));
  }, []);

  useEffect(refresh, [refresh]);

  const add = async () => {
    if (!name) return;
    try {
      await PolyToolsService.SetEnvVar(name, value, scope);
      setName(""); setValue("");
      refresh();
    } catch (e) {
      setError(msg(e));
    }
  };

  const del = async (v: EnvVar) => {
    try {
      await PolyToolsService.DeleteEnvVar(v.name, v.scope);
      refresh();
    } catch (e) {
      setError(msg(e));
    }
  };

  return (
    <>
      <div className="section-label">{t("panel.env", "Variables")}</div>
      <Card className="setting-group" size="small">
        <div className="setting-row">
          <div className="setting-text">
            <div className="setting-label">{t("panel.env.add", "Add / update variable")}</div>
            <div className="setting-desc">
              {t("panel.env.sysDesc", "System scope requires administrator rights.")}
            </div>
          </div>
          <div className="panel-row">
            <Input size="small" placeholder={t("panel.env.name", "Name")} value={name}
              onChange={(_, d) => setName(d.value)} style={{ width: 130 }} />
            <Input size="small" placeholder={t("panel.env.value", "Value")} value={value}
              onChange={(_, d) => setValue(d.value)} style={{ width: 200 }} />
            <Dropdown
              size="small" selectedOptions={[scope]}
              value={scope === "user" ? t("panel.env.user", "User") : t("panel.env.system", "System")}
              onOptionSelect={(_, d) => setScope(String(d.optionValue))}
              style={{ minWidth: 90 }}
            >
              <Option value="user">{t("panel.env.user", "User")}</Option>
              <Option value="system">{t("panel.env.system", "System")}</Option>
            </Dropdown>
            <Button size="small" appearance="primary" icon={<AddRegular />} onClick={add} />
          </div>
        </div>
        {error && <div className="panel-status panel-error">{error}</div>}
        <div className="panel-list">
          <Table size="small">
            <TableHeader>
              <TableRow>
                <TableHeaderCell>{t("panel.env.colName", "Name")}</TableHeaderCell>
                <TableHeaderCell>{t("panel.env.colValue", "Value")}</TableHeaderCell>
                <TableHeaderCell>{t("panel.env.colScope", "Scope")}</TableHeaderCell>
                <TableHeaderCell />
              </TableRow>
            </TableHeader>
            <TableBody>
              {vars.map((v) => (
                <TableRow key={v.scope + ":" + v.name}>
                  <TableCell>{v.name}</TableCell>
                  <TableCell className="panel-mono">{v.value}</TableCell>
                  <TableCell>{v.scope}</TableCell>
                  <TableCell>
                    <Button
                      size="small" appearance="subtle" icon={<DeleteRegular />}
                      onClick={() => del(v)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>
    </>
  );
}

// --- Hosts file editor ---

export function HostsPanel(_props: PanelProps) {
  const t = useT();
  const [content, setContent] = useState("");
  const [dirty, setDirty] = useState(false);
  const [status, setStatus] = useState("");

  useEffect(() => {
    PolyToolsService.ReadHosts()
      .then(setContent)
      .catch((e) => setStatus(msg(e)));
  }, []);

  const save = async () => {
    try {
      await PolyToolsService.WriteHosts(content);
      setDirty(false);
      setStatus(t("panel.hosts.saved", "Saved — DNS cache flushed."));
    } catch (e) {
      setStatus(msg(e));
    }
  };

  return (
    <>
      <div className="section-label">{t("panel.hosts", "hosts file")}</div>
      <Card className="setting-group" size="small">
        <div className="panel-hosts">
          <Textarea
            value={content} rows={14}
            onChange={(_, d) => { setContent(d.value); setDirty(true); }}
            className="panel-mono"
            style={{ width: "100%" }}
          />
        </div>
        <div className="panel-actions">
          {status && <span className="panel-status">{status}</span>}
          <Button appearance="primary" size="small" disabled={!dirty} onClick={save}>
            {t("panel.hosts.save", "Save (requires admin)")}
          </Button>
        </div>
      </Card>
    </>
  );
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

// --- Screensaver source picker ---

export function ScreensaverPanel({ module }: PanelProps) {
  const t = useT();
  const source = String(
    module.settings?.find((f) => f.key === "source")?.value ?? ""
  );
  const contentType = String(
    module.settings?.find((f) => f.key === "contentType")?.value ?? "video"
  );

  const pick = () => {
    const isFolder = contentType === "images";
    const fn = isFolder ? PolyToolsService.PickFolder : PolyToolsService.PickFile;
    fn.call(PolyToolsService, t("panel.ss.pick", "Select source"))
      .then((p) => {
        if (p) PolyToolsService.UpdateModuleSetting("screensaver", "source", p);
      })
      .catch(console.error);
  };

  return (
    <Card className="setting-group" size="small">
      <div className="setting-row">
        <div className="setting-text">
          <div className="setting-label">{t("panel.ss.current", "Current source")}</div>
          <div className="setting-desc">{source || t("panel.ss.none", "None selected")}</div>
        </div>
        <Button size="small" onClick={pick}>
          {contentType === "images"
            ? t("panel.ss.pickFolder", "Choose folder…")
            : t("panel.ss.pickFile", "Choose file…")}
        </Button>
      </div>
    </Card>
  );
}
