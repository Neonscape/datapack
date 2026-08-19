import {useEffect, useState} from "react";
import {GetProjectSettings, SelectProjectExportDirectory, SetProjectExportSettings} from "../../wailsjs/go/main/App";
import {workspace as workspaceModel} from "../../wailsjs/go/models";
import {useTranslation} from "../i18n";
import type {ProjectExportSettings, ProjectExportTable, ProjectSettings as ProjectSettingsModel} from "../types";

type Tab = "export" | "info";

export function ProjectSettings({onClose, onSaved}: {onClose: () => void; onSaved: (workspace: ProjectSettingsModel["workspace"]) => void}) {
    const {language, t} = useTranslation();
    const [tab, setTab] = useState<Tab>("export");
    const [settings, setSettings] = useState<ProjectSettingsModel | null>(null);
    const [tables, setTables] = useState<ProjectExportTable[]>([]);
    const [directory, setDirectory] = useState("");
    const [targets, setTargets] = useState<string[]>([]);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        void (async () => {
            try {
                const result = await GetProjectSettings() as ProjectSettingsModel;
                setSettings(result);
                setTables(result.tables);
                setDirectory(result.export.directory);
                setTargets(result.export.targets);
            } catch (reason) {
                setError(reason instanceof Error ? reason.message : String(reason));
            }
        })();
    }, []);

    function updateSelection(updater: (current: ProjectExportTable[]) => ProjectExportTable[]) {
        setTables((current) => updater(current));
    }

    function selectAll() {
        updateSelection((current) => current.map((table) => ({...table, selected: true})));
    }

    function clearAll() {
        updateSelection((current) => current.map((table) => ({...table, selected: false})));
    }

    function invertSelection() {
        updateSelection((current) => current.map((table) => ({...table, selected: !table.selected})));
    }

    function selectConfigured() {
        updateSelection((current) => current.map((table) => ({...table, selected: table.configured})));
    }

    async function selectDirectory() {
        setError("");
        try {
            const selected = await SelectProjectExportDirectory();
            if (selected) setDirectory(selected);
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    async function save() {
        setBusy(true);
        setError("");
        const payload: ProjectExportSettings = {
            directory,
            targets,
            tables: Object.fromEntries(tables.map((table) => [table.id, table.selected])),
        };
        try {
            const result = await SetProjectExportSettings(new workspaceModel.ProjectExportSettings(payload)) as ProjectSettingsModel;
            onSaved(result.workspace);
            onClose();
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        } finally {
            setBusy(false);
        }
    }

    return <div className="dialog-backdrop project-settings-backdrop" role="presentation" onMouseDown={onClose}>
        <form className="project-settings-dialog" onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); void save(); }}>
            <header className="project-settings-header">
                <div>
                    <h2>{t("projectSettings.title")}</h2>
                    {settings && <div className="project-settings-path">{settings.project.path}</div>}
                </div>
                <button type="button" className="peek-close" aria-label={t("common.close")} onClick={onClose}>×</button>
            </header>
            <nav className="project-settings-tabs" role="tablist">
                <button type="button" className={tab === "export" ? "project-settings-tab-active" : ""} onClick={() => setTab("export")}>{t("projectSettings.exportTab")}</button>
                <button type="button" className={tab === "info" ? "project-settings-tab-active" : ""} onClick={() => setTab("info")}>{t("projectSettings.infoTab")}</button>
            </nav>
            {tab === "export" ? <section className="project-settings-content">
                <label>{t("projectSettings.exportDirectory")}
                    <span className="project-settings-directory-input"><input value={directory} onChange={(event) => setDirectory(event.target.value)}/><button type="button" onClick={() => void selectDirectory()}>{t("projectSettings.chooseDirectory")}</button></span>
                </label>
                <div className="project-settings-section-title">{t("projectSettings.formats")}</div>
                <div className="project-settings-format-list">{(settings?.formats ?? []).map((format) => <label key={format.id}>
                    <input type="checkbox" checked={targets.includes(format.id)} onChange={() => setTargets((current) => current.includes(format.id) ? current.filter((target) => target !== format.id) : [...current, format.id])}/>
                    {format.labels[language] ?? format.labels.en ?? format.id}
                </label>)}</div>
                <div className="project-settings-section-title">{t("projectSettings.tables")}</div>
                <div className="project-settings-actions">
                    <button type="button" onClick={selectAll}>{t("projectSettings.selectAll")}</button>
                    <button type="button" onClick={clearAll}>{t("projectSettings.clearAll")}</button>
                    <button type="button" onClick={invertSelection}>{t("projectSettings.invert")}</button>
                    <button type="button" onClick={selectConfigured}>{t("projectSettings.selectConfigured")}</button>
                </div>
                <div className="project-settings-summary">{t("projectSettings.selectedSummary", {selected: tables.filter((table) => table.selected).length, total: tables.length})}</div>
                <div className="project-settings-table-list">
                    {tables.map((table) => <label className="project-settings-table-row" key={table.id}>
                        <input type="checkbox" checked={table.selected} onChange={() => updateSelection((current) => current.map((entry) => entry.id === table.id ? {...entry, selected: !entry.selected} : entry))}/>
                        <span className="project-settings-table-main"><strong>{table.id}</strong><small>{table.path}</small></span>
                        <span className={table.configured ? "project-settings-configured" : "project-settings-unconfigured"}>{table.configured ? table.columns?.join(", ") : t("projectSettings.notConfigured")}</span>
                    </label>)}
                    {tables.length === 0 && <div className="schema-editor-empty">{t("projectSettings.noTables")}</div>}
                </div>
            </section> : <section className="project-settings-content project-settings-info">
                <div><strong>{t("projectSettings.projectPath")}</strong><code>{settings?.project.path}</code></div>
                <div><strong>{t("projectSettings.projectRoot")}</strong><code>{settings?.project.root}</code></div>
                <div><strong>{t("projectSettings.tableRoots")}</strong><code>{settings?.project.tableRoots.join(", ")}</code></div>
                <div><strong>{t("projectSettings.schemaRoots")}</strong><code>{settings?.project.schemaRoots.join(", ")}</code></div>
                <div><strong>{t("projectSettings.exclude")}</strong><code>{settings?.project.exclude?.join(", ") || "-"}</code></div>
            </section>}
            {error && <div className="schema-editor-error">{error}</div>}
            {tab === "export" && <div className="dialog-actions"><button type="button" onClick={onClose}>{t("common.cancel")}</button><button type="submit" disabled={busy}>{t("common.apply")}</button></div>}
        </form>
    </div>;
}
