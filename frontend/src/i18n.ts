import i18n from "i18next";
import { initReactI18next } from "react-i18next";

import en from "@locales/en/translation.json";
import vi from "@locales/vi/translation.json";
import ja from "@locales/ja/translation.json";
import zh from "@locales/zh/translation.json";
import ko from "@locales/ko/translation.json";

// Default application language is English ("en")
const getInitialLanguage = (): string => {
  try {
    const stored = localStorage.getItem("neko_app_lang") || localStorage.getItem("i18nextLng");
    // If stored language was Vietnamese or missing, migrate to English default
    if (stored && stored !== "vi" && ["en", "ja", "zh", "ko"].includes(stored)) {
      return stored;
    }
    localStorage.setItem("neko_app_lang", "en");
    localStorage.setItem("i18nextLng", "en");
  } catch {
    // ignore
  }
  return "en";
};

const initialLng = getInitialLanguage();

i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    vi: { translation: vi },
    ja: { translation: ja },
    zh: { translation: zh },
    ko: { translation: ko },
  },
  lng: initialLng,
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
