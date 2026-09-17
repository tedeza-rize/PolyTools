import {
  AppsRegular,
  ArrowMoveRegular,
  BracesVariableRegular,
  ColorRegular,
  CursorRegular,
  EyeRegular,
  GridRegular,
  GlobeRegular,
  KeyboardRegular,
  PinRegular,
  RenameRegular,
  RulerRegular,
  TextFontRegular,
  WeatherMoonRegular,
  WindowMultipleRegular,
} from "@fluentui/react-icons";

const ICONS: Record<string, JSX.Element> = {
  Pin: <PinRegular />,
  WeatherMoon: <WeatherMoonRegular />,
  TextFont: <TextFontRegular />,
  Color: <ColorRegular />,
  Ruler: <RulerRegular />,
  ArrowMove: <ArrowMoveRegular />,
  WindowMultiple: <WindowMultipleRegular />,
  Grid: <GridRegular />,
  Keyboard: <KeyboardRegular />,
  Cursor: <CursorRegular />,
  Rename: <RenameRegular />,
  Eye: <EyeRegular />,
  BracesVariable: <BracesVariableRegular />,
  Globe: <GlobeRegular />,
};

export function ModuleIcon({ name }: { name: string }) {
  return <>{ICONS[name] ?? <AppsRegular />}</>;
}
