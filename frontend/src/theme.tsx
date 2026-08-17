import {createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode} from "react";
import type {Theme as DataGridTheme} from "@glideapps/glide-data-grid";

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
    dataGridTheme: Partial<DataGridTheme> | undefined;
    setPreference: (preference: ThemePreference) => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

const gridPalette: Record<ResolvedTheme, GridColors> = {
    light: {errorCellBg: "#fff1f1", errorCellText: "#8a3030", draftCellBg: "#f7fafb", draftCellText: "#51616b", formulaBg: "#e8f3f5", formulaText: "#1c5d73"},
    dark: {errorCellBg: "#3a2023", errorCellText: "#e08a8a", draftCellBg: "#1d2931", draftCellText: "#aab6bd", formulaBg: "#173039", formulaText: "#7fd0dc"},
};

// Overrides the glide-data-grid default (light) theme when the app is dark.
const darkDataGridTheme: Partial<DataGridTheme> = {
    accentColor: "#2f9db0",
    accentFg: "#ffffff",
    accentLight: "rgba(47, 157, 176, 0.15)",
    textDark: "#d5dde3",
    textMedium: "#8fa0ab",
    textLight: "#6b7b85",
    textBubble: "#d5dde3",
    bgIconHeader: "#8fa0ab",
    fgIconHeader: "#12181d",
    textHeader: "#d5dde3",
    textGroupHeader: "#d5dde3bb",
    textHeaderSelected: "#ffffff",
    bgCell: "#1b242b",
    bgCellMedium: "#202a32",
    bgHeader: "#202a32",
    bgHeaderHasFocus: "#2a3640",
    bgHeaderHovered: "#23313b",
    bgBubble: "#23313b",
    bgBubbleSelected: "#2a3640",
    bgSearchResult: "#332b1a",
    borderColor: "rgba(138, 160, 171, 0.16)",
    drilldownBorder: "rgba(47, 157, 176, 0.4)",
    linkColor: "#7fd0dc",
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
    const dataGridTheme = resolved === "dark" ? darkDataGridTheme : undefined;

    const value = useMemo(() => ({preference, resolved, grid, dataGridTheme, setPreference}), [preference, resolved, grid, dataGridTheme, setPreference]);
    return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
    const context = useContext(ThemeContext);
    if (!context) throw new Error("useTheme must be used within ThemeProvider");
    return context;
}
