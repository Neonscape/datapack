import {useState} from 'react';
import './App.css';
import {SelectProject} from "../wailsjs/go/main/App";

type TreeNode = {
    kind: "project" | "group" | "directory" | "tableFile" | "schemaFile";
    name: string;
    path?: string;
    children?: TreeNode[];
};

type WorkspaceView = {
    project: {
        path: string;
        root: string;
    };
    settings: {
        build: {
            mode: "auto" | "manual";
            autoCompileDebounceSeconds?: number;
        };
    };
    tree: {
        root: TreeNode;
    };
};

function Tree({node, depth = 0}: { node: TreeNode; depth?: number }) {
    const isFile = node.kind === "tableFile" || node.kind === "schemaFile";
    return (
        <div className="tree-node">
            <div className={`tree-row ${isFile ? "tree-file" : ""}`} style={{paddingLeft: `${12 + depth * 16}px`}}>
                <span className={`node-icon node-${node.kind}`} aria-hidden="true"/>
                <span>{node.name}</span>
            </div>
            {node.children?.map((child) => <Tree key={`${child.kind}:${child.path ?? child.name}`} node={child} depth={depth + 1}/>) }
        </div>
    );
}

function App() {
    const [workspace, setWorkspace] = useState<WorkspaceView | null>(null);
    const [error, setError] = useState("");

    async function selectProject() {
        setError("");
        try {
            const view = await SelectProject() as WorkspaceView;
            if (view?.project?.path) {
                setWorkspace(view);
            }
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : String(reason));
        }
    }

    return (
        <main className="app-shell">
            <header className="app-header">
                <div className="brand">Datapack</div>
                <div className="header-divider"/>
                <span className="project-name">{workspace?.tree.root.name ?? "No project open"}</span>
                <div className="header-actions">
                    {workspace && <span className={`build-mode build-mode-${workspace.settings.build.mode}`}>{workspace.settings.build.mode === "auto" ? "Auto compile" : "Manual compile"}</span>}
                    <button className="open-button" onClick={selectProject}>Open project</button>
                </div>
            </header>

            {workspace ? (
                <section className="workbench">
                    <aside className="project-explorer">
                        <div className="panel-title">Project</div>
                        <div className="tree"><Tree node={workspace.tree.root}/></div>
                    </aside>
                    <section className="editor-placeholder">
                        <h1>Select a source file</h1>
                        <p>The project tree is ready. Table and Schema editors will open here.</p>
                    </section>
                </section>
            ) : (
                <section className="welcome">
                    <div className="welcome-mark">DP</div>
                    <h1>Open a data project</h1>
                    <p>Choose a <code>.dtproj</code> file to inspect its tables and public Schema files.</p>
                    <button className="open-button open-button-large" onClick={selectProject}>Open project</button>
                </section>
            )}
            {error && <div className="error-toast" role="alert">{error}</div>}
        </main>
    )
}

export default App
