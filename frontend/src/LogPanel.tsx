import {useEffect, useState} from "react";
import {EventsOn} from "../wailsjs/runtime/runtime";
import {useTranslation} from "./i18n";
import type {LogEntry} from "./types";

export function LogPanel() {
    const {t} = useTranslation();
    const [entries, setEntries] = useState<LogEntry[]>([]);
    const [open, setOpen] = useState(false);
    useEffect(() => {
        const append = (entry: LogEntry) => {
            setEntries((current) => [...current, entry].slice(-500));
        };
        if ((window as Window & {runtime?: {EventsOn?: unknown}}).runtime?.EventsOn) {
            return EventsOn("log:entry", append);
        }
        const fallback = (event: Event) => append((event as CustomEvent<LogEntry>).detail);
        window.addEventListener("log:entry", fallback);
        return () => window.removeEventListener("log:entry", fallback);
    }, []);
    return <section className={`log-panel ${open ? "log-panel-open" : ""}`}>
        <button className="log-panel-toggle" onClick={() => setOpen(!open)} aria-expanded={open}>
            {t("log.title")} <span>{entries.length}</span>
        </button>
        {open && <div className="log-panel-body">
            <div className="log-panel-toolbar"><span>{t("log.applicationLog")}</span><button onClick={() => setEntries([])}>{t("log.clear")}</button></div>
            <div className="log-list">
                {entries.map((entry, index) => <div className={`log-entry log-${entry.level}`} key={`${entry.time}-${index}`}>
                    <time>{new Date(entry.time).toLocaleTimeString()}</time>
                    <strong>{entry.level.toUpperCase()}</strong>
                    <span>{entry.message}</span>
                    {entry.fields && <code>{JSON.stringify(entry.fields)}</code>}
                </div>)}
                {entries.length === 0 && <div className="log-empty">{t("log.empty")}</div>}
            </div>
        </div>}
    </section>;
}
