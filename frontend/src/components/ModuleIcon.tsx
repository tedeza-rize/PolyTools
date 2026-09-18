import {
  AppsRegular,
  ArrowMaximizeRegular,
  BatteryChargeRegular,
  BorderNoneRegular,
  FlowRegular,
  GaugeRegular,
  KeyCommandRegular,
  MoreHorizontalRegular,
  PinRegular,
  TvRegular,
  WeatherMoonRegular,
} from "@fluentui/react-icons";

// Distinct accent color per module, like the multicolored icons in
// Windows 11 Settings' navigation.
const COLORS: Record<string, string> = {
  Pin: "#0f6cbd",
  WeatherMoon: "#5b5fc7",
  BatteryCharge: "#107c10",
  Tv: "#0f6cbd",
  BorderNone: "#107c10",
  Gauge: "#d13438",
  ArrowMaximize: "#008272",
  KeyCommand: "#5b5fc7",
  MoreHorizontal: "#69797e",
  Flow: "#5b5fc7",
};

export function moduleColor(name: string): string {
  return COLORS[name] ?? "#0f6cbd";
}

const ICONS: Record<string, JSX.Element> = {
  Pin: <PinRegular />,
  WeatherMoon: <WeatherMoonRegular />,
  BatteryCharge: <BatteryChargeRegular />,
  Tv: <TvRegular />,
  BorderNone: <BorderNoneRegular />,
  Gauge: <GaugeRegular />,
  ArrowMaximize: <ArrowMaximizeRegular />,
  KeyCommand: <KeyCommandRegular />,
  MoreHorizontal: <MoreHorizontalRegular />,
  Flow: <FlowRegular />,
};

export function ModuleIcon({ name }: { name: string }) {
  return <>{ICONS[name] ?? <AppsRegular />}</>;
}
