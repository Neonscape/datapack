import {useRef} from "react";
import {useTranslation} from "../i18n";
import {useModalBoundary} from "../project/useModalBoundary";

export function DisplayNameDialog({identity, value, onChange, onClose, onSubmit}: {
    identity: string; value: string; onChange: (value: string) => void;
    onClose: () => void; onSubmit: () => void;
}) {
    const {t} = useTranslation();
    const backdrop = useRef<HTMLDivElement>(null);
    useModalBoundary(backdrop, onClose);
    return <div ref={backdrop} className="dialog-backdrop" role="presentation" onMouseDown={onClose}>
        <form className="structure-dialog" role="dialog" aria-modal="true" aria-labelledby="display-name-title" tabIndex={-1}
            onMouseDown={(event) => event.stopPropagation()} onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
            <h2 id="display-name-title">{t("displayName.edit")}</h2>
            <p>{identity}</p>
            <label>{t("displayName.label")}<input autoFocus value={value} onChange={(event) => onChange(event.target.value)}/></label>
            <p>{t("displayName.hint")}</p>
            <div className="dialog-actions"><button type="button" onClick={onClose}>{t("common.cancel")}</button><button type="submit">{t("common.apply")}</button></div>
        </form>
    </div>;
}
