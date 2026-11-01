import {useEffect, useRef, type RefObject} from "react";

// Lock the workspace, while allowing native wheel scrolling inside the dialog.
// Body overflow alone does not contain scrollable grid/side-panel elements.
export function useModalBoundary(ref: RefObject<HTMLElement>, onClose: () => void) {
    const close = useRef(onClose);
    close.current = onClose;
    useEffect(() => {
        const backdrop = ref.current;
        if (!backdrop) return;
        const dialog = backdrop.querySelector<HTMLElement>('[role="dialog"]');
        if (!dialog) return;
        const previousFocus = document.activeElement as HTMLElement | null;
        const siblings = Array.from(backdrop.parentElement?.children ?? [])
            .filter((element): element is HTMLElement => element instanceof HTMLElement && element !== backdrop);
        const inertStates = siblings.map((element) => element.inert);
        siblings.forEach((element) => { element.inert = true; });
        const bodyOverflow = document.body.style.overflow;
        document.body.style.overflow = "hidden";
        dialog.focus();

        const wheel = (event: WheelEvent) => {
            event.stopPropagation();
            let target = event.target instanceof Element ? event.target : null;
            let canScroll = false;
            while (target && dialog.contains(target)) {
                const style = getComputedStyle(target);
                const vertical = /auto|scroll/.test(style.overflowY) &&
                    (event.deltaY < 0 ? target.scrollTop > 0 : event.deltaY > 0 && target.scrollTop + target.clientHeight < target.scrollHeight);
                const horizontal = /auto|scroll/.test(style.overflowX) &&
                    (event.deltaX < 0 ? target.scrollLeft > 0 : event.deltaX > 0 && target.scrollLeft + target.clientWidth < target.scrollWidth);
                canScroll ||= vertical || horizontal;
                if (target === dialog) break;
                target = target.parentElement;
            }
            if (!canScroll) event.preventDefault();
        };
        const keydown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                event.preventDefault();
                event.stopPropagation();
                close.current();
            } else if (event.key === "Tab") {
                const controls = Array.from(dialog.querySelectorAll<HTMLElement>('button:not(:disabled), input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex="0"]'))
                    .filter((element) => element.getClientRects().length > 0);
                const first = controls[0];
                const last = controls.at(-1);
                if (!first || (event.shiftKey && (document.activeElement === first || document.activeElement === dialog))) {
                    event.preventDefault();
                    (last ?? dialog).focus();
                } else if (!event.shiftKey && (document.activeElement === last || !dialog.contains(document.activeElement))) {
                    event.preventDefault();
                    first.focus();
                }
            }
        };
        window.addEventListener("wheel", wheel, {capture: true, passive: false});
        window.addEventListener("keydown", keydown, true);
        return () => {
            window.removeEventListener("wheel", wheel, true);
            window.removeEventListener("keydown", keydown, true);
            siblings.forEach((element, index) => { element.inert = inertStates[index]; });
            document.body.style.overflow = bodyOverflow;
            if (previousFocus?.isConnected && !previousFocus.closest("[inert]")) previousFocus.focus();
        };
    }, [ref]);
}
