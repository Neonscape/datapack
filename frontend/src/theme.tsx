import {createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode} from "react";

export type ThemePreference = "system" | "light" | "dark";
export type ResolvedTheme = "light" | "dark";

export type GridColors = {
    errorCellBg: string;
    errorCellText: string;
    draftCellBg: string;
    draftCellText: string;
    formulaBg: string;
    formulaText: string;
};

type ThemeContextValue = {
    preference: ThemePreference;
    resolved: ResolvedTheme;
    grid: GridColors;
    setPreference: (preference: ThemePreference) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

const gridPalette: Record<ResolvedTheme, GridColors> = {
    light: {errorCellBg: "#fff1f1", errorCellText: "#8a3030", draftCellBg: "#f7fafb", draftCellText: "#51616b", formulaBg: "#e8f3f5", formulaText: "#1c5d73"},
    dark: {errorCellBg: "#3a2023", errorCellText: "#e08a8a", draftCellBg: "#1d2931", draftCellText: "#aab6bd", formulaBg: "#173039", formulaText: "#7fd0dc"},
};

function systemTheme(): ResolvedTheme {
    if (typeof window === "undefined" || !window.matchMedia) return "light";
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function ThemeProvider({children}: {children: ReactNode}) {
    const [preference, setPreferenceState] = useState<ThemePreference>("system");
    const [system, setSystem] = useState<ResolvedTheme>(() => systemTheme());
    const resolved: ResolvedTheme = preference === "system" ? system : preference;

    useEffect(() => {
        document.documentElement.dataset.theme = resolved;
    }, [resolved]);

    useEffect(() => {
        if (typeof window === "undefined" || !window.matchMedia) return;
        const query = window.matchMedia("(prefers-color-scheme: dark)");
        const onChange = (event: MediaQueryListEvent) => setSystem(event.matches ? "dark" : "light");
        query.addEventListener("change", onChange);
        return () => query.removeEventListener("change", onChange);
    }, []);

    const setPreference = useCallback((next: ThemePreference) => setPreferenceState(next), []);
    const grid = gridPalette[resolved];

    const value = useMemo(() => ({preference, resolved, grid, setPreference}), [preference, resolved, grid, setPreference]);
    return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
    const context = useContext(ThemeContext);
    if (!context) throw new Error("useTheme must be used within ThemeProvider");
    return context;
}
