import { defineStore } from "pinia";
import { ref } from "vue";
import { getEffectiveLocale, saveStoreConfig } from "@/api/http";
import { applyLocale, guessInitialLocale, type LocaleChoice, type LocaleCode } from "@/i18n";

/**
 * 界面语言 Store。
 *
 * - `choice`：用户的显式选择（"auto" 表示交给后端检测），通过 persist 插件持久化到
 *   后端配置文件（config["locale"]）。
 * - `effective` / `source`：后端 GET /api/locale 解析出的最终生效语言（瞬态字段，不持久化）。
 *
 * 启动流程：应用挂载前调用 `apply()`，由后端决定默认语言（用户持久化选择优先，
 * 否则按 Accept-Language 检测，失败回退英语）。
 */
export const useLocaleStore = defineStore(
  "locale",
  () => {
    const choice = ref<LocaleChoice>("auto");
    const effective = ref<LocaleCode>(guessInitialLocale());
    const source = ref<"user" | "detected">("detected");

    /** 向后端请求生效语言并应用到 vue-i18n。 */
    async function apply() {
      const result = await getEffectiveLocale();
      if (!result) return;
      const locale = normalizeLocale(result.locale);
      effective.value = locale;
      source.value = result.source;
      applyLocale(locale);
    }

    /** 用户切换语言：先持久化到后端配置，再应用语言，最后 reload 一次页面避免渲染问题。 */
    async function setChoice(next: LocaleChoice) {
      choice.value = next;

      // 显式保存（不走 persist 插件的防抖），确保 reload 前配置已落盘；
      // 后端内存配置同步更新，随后 apply() 即可读到新选择。
      await saveStoreConfig("locale", { choice: next });

      await apply();

      // 变更界面语言后 reload 一次，避免既有组件（如 Monaco、nuxt-ui 内部缓存）
      // 残留旧语言渲染。reload 后 main.ts 会重新从后端解析语言。
      window.location.reload();
    }

    return { choice, effective, source, apply, setChoice };
  },
  { persist: true, persistExclude: ["effective", "source"] },
);

/** 将后端返回的 locale 归一化为受支持的语言码，未知时回退英语。 */
function normalizeLocale(locale: string): LocaleCode {
  return locale === "zh-CN" ? "zh-CN" : "en";
}
