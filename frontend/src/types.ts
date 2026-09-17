// Re-export the generated binding models so the app uses one source of truth.
export {
  Category,
  SettingType,
  type General,
  type Info,
  type SelectOption,
  type SettingField,
} from "../bindings/polytools/internal/core/models";

import { Category, type Info } from "../bindings/polytools/internal/core/models";

export type ModuleInfo = Info;

export const CATEGORY_ORDER: Category[] = [
  Category.CategorySystem,
  Category.CategoryWindowing,
  Category.CategoryInput,
  Category.CategoryFiles,
  Category.CategoryAdvanced,
];

export const CATEGORY_LABELS: Record<Category, string> = {
  [Category.$zero]: "",
  [Category.CategorySystem]: "System Tools",
  [Category.CategoryWindowing]: "Windowing & Layouts",
  [Category.CategoryInput]: "Input / Output",
  [Category.CategoryFiles]: "File Management",
  [Category.CategoryAdvanced]: "Advanced",
};
