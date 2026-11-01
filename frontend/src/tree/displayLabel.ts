// A title is presentation only. Always retain the internal name for disambiguation.
export function displayLabel(title: string | undefined, identity: string): string {
    const normalized = title?.trim();
    return normalized ? `${normalized} (${identity})` : identity;
}
