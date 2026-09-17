import {
  Dropdown,
  Input,
  Option,
  Slider,
  Switch,
} from "@fluentui/react-components";
import { SettingField, SettingType } from "../types";

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
      return <kbd className="hotkey">{formatShortcut(String(field.value))}</kbd>;
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
