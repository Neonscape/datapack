// Mounted inactive tabs retain drafts/selections but cannot own a document
// action. Native inputs/contenteditable retain their own clipboard/navigation.
export function ownsGridEvent(active:boolean,element:HTMLElement|null,event:Event,focused:Element|null):boolean {
    if(!active || !element || event.defaultPrevented || !element.contains(focused))return false;
    const target=event.target;
    return target instanceof Element && !target.closest("textarea, input, select") && !(target instanceof HTMLElement && target.isContentEditable);
}
