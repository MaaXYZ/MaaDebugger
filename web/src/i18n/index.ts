import { createI18n } from "vue-i18n";
import en from "./locales/en";
import zhCN from "./locales/zh-CN";

/** 已注册的界面语言。新增语言时在此追加即可（同时需同步后端 locale.go 的支持列表）。 */
export const localeOptions = [
  { code: "en", label: "English" },
  { code: "zh-CN", label: "简体中文" },
] as const;

export type LocaleCode = (typeof localeOptions)[number]["code"];
/** 用户可选的界面语言：具体语言，或 "auto"（由后端检测决定）。 */
export type LocaleChoice = LocaleCode | "auto";

export const messages = {
  en,
  "zh-CN": zhCN,
} as const;

export type MessageSchema = (typeof messages)["en"];

declare module "vue-i18n" {
  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  export interface DefineLocaleMessage extends MessageSchema {}
}

export const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackLocale: "en",
  messages,
});

/**
 * 应用界面语言：设置 vue-i18n locale 与 <html lang>。
 */
export function applyLocale(locale: LocaleCode) {
  i18n.global.locale.value = locale;
  document.documentElement.lang = locale;
}

/**
 * 挂载前的乐观初始语言：根据浏览器语言粗略估计，避免首屏闪烁。
 * 真正的默认语言由后端（GET /api/locale）决定，会在应用启动时覆盖此值。
 */
export function guessInitialLocale(): LocaleCode {
  const nav = typeof navigator !== "undefined" ? navigator.language || "" : "";
  return nav.toLowerCase().startsWith("zh") ? "zh-CN" : "en";
}
