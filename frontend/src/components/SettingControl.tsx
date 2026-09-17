import { useEffect, useState } from "react";
import {
  Button,
  Dropdown,
  Input,
  Option,
  Slider,
  Switch,
} from "@fluentui/react-components";
import { SettingField, SettingType } from "../types";
import { useT } from "../i18n";

const MODIFIER_LABELS: Record<string, string> = {
  super: "Win",
  ctrl: "Ctrl",
  alt: "Alt",
  shift: "Shift",
  cmdorctrl: "Ctrl",
};

export function formatShortcut(shortcut: string): string {
  return shortcut
    .split("+")
    .map((p) => {
      const lower = p.trim().toLowerCase();
      return MODIFIER_LABELS[lower] ?? p.trim().toUpperCase();
    })
    .join(" + ");
}

// Maps DOM KeyboardEvent.key values to Wails accelerator names.
const KEY_MAP: Record<string, string> = {
  " ": "space",
  Spacebar: "space",
  Enter: "enter",
  Return: "enter",
  Backspace: "backspace",
  Tab: "tab",
  Delete: "delete",
  ArrowLeft: "left",
  ArrowRight: "right",
  ArrowUp: "up",
  ArrowDown: "down",
  Home: "home",
  End: "end",
  PageUp: "page up",
  PageDown: "page down",
  NumLock: "numlock",
  "+": "plus",
};

const MODIFIER_KEYS = new Set(["Control", "Alt", "Shift", "Meta", "CapsLock", "AltGraph"]);

function keyName(k: string): string | null {
  if (MODIFIER_KEYS.has(k)) return null;
  if (/^F\d{1,2}$/.test(k)) return k.toLowerCase();
  const mapped = KEY_MAP[k];
  if (mapped) return mapped;
  if (k.length === 1) return k.toLowerCase();
  return null;
}

function HotkeyCapture({
  value,
  disabled,
  onChange,
}: {
  value: string;
  disabled?: boolean;
  onChange: (value: string) => void;
}) {
  const t = useT();
  const [capturing, setCapturing] = useState(false);

  useEffect(() => {
    if (!capturing) return;
    const onKey = (e: KeyboardEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (e.key === "Escape") {
        setCapturing(false);
        return;
      }
      const key = keyName(e.key);
      if (!key) return; // still only holding modifiers
      const mods: string[] = [];
      if (e.ctrlKey) mods.push("ctrl");
      if (e.altKey) mods.push("alt");
      if (e.shiftKey) mods.push("shift");
      if (e.metaKey) mods.push("super");
      onChange([...mods, key].join("+"));
      setCapturing(false);
    };
    window.addEventListener("keydown", onKey, true);
    return () => window.removeEventListener("keydown", onKey, true);
  }, [capturing, onChange]);

  return (
    <Button
      size="small"
      className="hotkey-btn"
      appearance={capturing ? "primary" : "outline"}
      disabled={disabled}
      onClick={() => setCapturing((c) => !c)}
    >
      {capturing ? t("hotkey.press") : formatShortcut(value)}
    </Button>
  );
}

export function SettingControl({
  field,
  disabled,
  onChange,
}: {
  field: SettingField;
  disabled?: boolean;
  onChange: (value: any) => void;
}) {
  switch (field.type) {
    case SettingType.SettingToggle:
      return (
        <Switch
          checked={!!field.value}
          disabled={disabled}
          onChange={(_, d) => onChange(d.checked)}
        />
      );
    case SettingType.SettingSelect:
      return (
        <Dropdown
          size="small"
          disabled={disabled}
          selectedOptions={[String(field.value)]}
          value={
            field.options?.find((o) => o.value === field.value)?.label ??
            String(field.value)
          }
          onOptionSelect={(_, d) => onChange(d.optionValue)}
          style={{ minWidth: 160 }}
        >
          {field.options?.map((o) => (
            <Option key={o.value} value={o.value}>
              {o.label}
            </Option>
          ))}
        </Dropdown>
      );
    case SettingType.SettingSlider:
      return (
        <Slider
          size="small"
          disabled={disabled}
          value={Number(field.value)}
          min={field.min ?? 0}
          max={field.max ?? 100}
          step={field.step ?? 1}
          onChange={(_, d) => onChange(d.value)}
          style={{ width: 180 }}
        />
      );
    case SettingType.SettingShortcut:
      return (
        <HotkeyCapture
          value={String(field.value)}
          disabled={disabled}
          onChange={onChange}
        />
      );
    case SettingType.SettingText:
    default:
      return (
        <Input
          size="small"
          disabled={disabled}
          value={String(field.value ?? "")}
          onChange={(_, d) => onChange(d.value)}
          style={{ minWidth: 200 }}
        />
      );
  }
}
