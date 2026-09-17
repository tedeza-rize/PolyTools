import { createContext, useContext } from "react";
import { en } from "./en";
import { ko } from "./ko";

export type Language = "en" | "ko";

export type Translator = (key: string, fallback?: string) => string;

const dicts: Record<Language, Record<string, string>> = { en, ko };

export function resolveLanguage(setting: string): Language {
  if (setting === "ko" || setting === "en") return setting;
  return (navigator.language || "en").toLowerCase().startsWith("ko")
    ? "ko"
    : "en";
}

export function translatorFor(setting: string): Translator {
  const lang = resolveLanguage(setting);
  return (key, fallback) => dicts[lang][key] ?? dicts.en[key] ?? fallback ?? key;
}

const I18nContext = createContext<Translator>((key, fallback) => fallback ?? key);

export const I18nProvider = I18nContext.Provider;

export function useT(): Translator {
  return useContext(I18nContext);
}
