import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import LanguageDetector from "i18next-browser-languagedetector"

import zh from "./locales/zh.json"
import en from "./locales/en.json"

const resources = {
  zh: { translation: zh },
  en: { translation: en },
}

i18n.use(initReactI18next)

// Keep tests deterministic: avoid browser language detection in Vitest/jsdom.
const isTest = import.meta.env.MODE === "test"
if (!isTest) {
  i18n.use(LanguageDetector)
}

i18n
  .init({
    resources,
    supportedLngs: ["zh", "en"],
    fallbackLng: "zh",
    lng: isTest ? "zh" : undefined,
    interpolation: {
      escapeValue: false,
    },
    detection: isTest
      ? undefined
      : {
          order: ["localStorage", "navigator"],
          caches: ["localStorage"],
        },
  })

export default i18n
