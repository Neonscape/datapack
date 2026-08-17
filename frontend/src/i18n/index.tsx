import {createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode} from "react";
import {en} from "./en";
import {zh} from "./zh";

export type Language = "en" | "zh-CN";

const catalogs: Record<Language, Record<string, string>> = {en, "zh-CN": zh};

type I18nContextValue = {
    language: Language;
    setLanguage: (language: Language) => void;
    t: (key: string, params?: Record<string, string | number>) => string;
};

const I18nContext = createContext<I18nContextValue | null>(null);

function interpolate(message: string, params?: Record<string, string | number>): string {
    if (!params) return message;
    return message.replace(/\{(\w+)\}/g, (match, key) => (key in params ? String(params[key]) : match));
}

export function I18nProvider({children}: {children: ReactNode}) {
    const [language, setLanguageState] = useState<Language>("zh-CN");

    useEffect(() => {
        document.documentElement.lang = language;
    }, [language]);

    const t = useCallback((key: string, params?: Record<string, string | number>) => {
        const catalog = catalogs[language] ?? catalogs["zh-CN"];
        const message = catalog[key] ?? catalogs["zh-CN"][key] ?? key;
        return interpolate(message, params);
    }, [language]);

    const setLanguage = useCallback((next: Language) => setLanguageState(next), []);
    const value = useMemo(() => ({language, setLanguage, t}), [language, setLanguage, t]);
    return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useTranslation(): I18nContextValue {
    const context = useContext(I18nContext);
    if (!context) throw new Error("useTranslation must be used within I18nProvider");
    return context;
}
