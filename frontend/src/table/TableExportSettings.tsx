import {useEffect, useState} from "react";
import {useTranslation} from "../i18n";
import type {ExportFormat, TableDocument, TableExportSettings as Settings} from "../types";

export function TableExportSettings({table, formats, onSave}: {table: TableDocument; formats: ExportFormat[]; onSave: (settings: Settings) => Promise<void>}) {
    const {language, t} = useTranslation();
    const [open, setOpen] = useState(false);
    const [settings, setSettings] = useState<Settings["formats"]>({});
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const columns = [{name: "id", label: "ID"}, ...table.columns.map((column) => ({name: column.name, label: column.displayName ? `${column.displayName} (${column.name})` : column.name}))];

    useEffect(() => {
        const next: Settings["formats"] = {};
        for (const format of formats) {
            const source = table.exports?.[format.id];
            next[format.id] = {enabled: source?.enabled ?? false, columns: source?.columns ?? []};
        }
        setSettings(next);
        setError("");
    }, [table, formats]);

    function toggleFormat(format: ExportFormat) {
        setSettings((current) => {
            const entry = current[format.id] ?? {enabled: false, columns: []};
            const enabled = !entry.enabled;
            return {...current, [format.id]: {enabled, columns: enabled && format.supportsColumns && (entry.columns?.length ?? 0) === 0 ? columns.map((column) => column.name) : entry.columns ?? []}};
        });
    }

    function setColumns(id: string, columns: string[]) {
        setSettings((current) => ({...current, [id]: {...current[id], columns}}));
    }

    async function save() {
        setBusy(true);
        setError("");
        try {
            await onSave({path: table.path, formats: settings});
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        } finally {
            setBusy(false);
        }
    }

    return <section className={`table-export-settings ${open ? "table-export-settings-open" : ""}`}>
        <button className="table-export-settings-toggle" type="button" aria-expanded={open} onClick={() => setOpen(!open)}>
            <span>{t("tableExport.title")}</span>
            <span>{open ? "−" : "+"}</span>
        </button>
        {open && <div className="table-export-settings-body">
            <div className="table-export-settings-formats">
                {formats.map((format) => <label className="table-export-format-toggle" key={format.id}>
                    <input type="checkbox" checked={!!settings[format.id]?.enabled} onChange={() => toggleFormat(format)}/>
                    <span>{format.labels[language] ?? format.labels.en ?? format.id}</span>
                </label>)}
            </div>
            {formats.filter((format) => settings[format.id]?.enabled).map((format) => <section className="table-export-format-config" key={format.id}>
                <h2>{format.labels[language] ?? format.labels.en ?? format.id} {t("tableExport.configuration")}</h2>
                {format.supportsColumns && <>
                    <div className="table-export-columns-head">
                        <span>{t("tableExport.columns")}</span>
                        <div>
                            <button type="button" onClick={() => setColumns(format.id, columns.map((column) => column.name))}>{t("tableExport.selectAllColumns")}</button>
                            <button type="button" onClick={() => setColumns(format.id, [])}>{t("tableExport.clearColumns")}</button>
                        </div>
                    </div>
                    <div className="table-export-columns">{columns.map((column) => <label key={column.name}>
                        <input type="checkbox" checked={settings[format.id]?.columns?.includes(column.name) ?? false} onChange={() => {
                            const selected = settings[format.id]?.columns ?? [];
                            setColumns(format.id, selected.includes(column.name) ? selected.filter((entry) => entry !== column.name) : [...selected, column.name]);
                        }}/>
                        {column.label}
                    </label>)}</div>
                </>}
            </section>)}
            {formats.length === 0 && <div className="schema-editor-empty">{t("tableExport.noFormats")}</div>}
            {error && <div className="schema-editor-error">{error}</div>}
            <div className="table-export-actions"><button type="button" disabled={busy} onClick={() => void save()}>{t("common.apply")}</button></div>
        </div>}
    </section>;
}
