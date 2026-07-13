export const hostnameRegex = /^(PCRM|RM)\d+$/;

export const hostnameWhitelist = [
    "STRAN1-PC",
    "COSE1-PC",
    "RM094",
    "FOISX-PC"
]

export const devHostnameWhitelist = [
    "DESKTOP-7PDPATV",
    "FOISX-PC"
]

export const isDevMachine = (hostname: string, adDomain: string): boolean => {
    if (devHostnameWhitelist.includes(hostname)) {
        if(hostname === "DESKTOP-7PDPATV" && adDomain === "WORKGROUP") {
            return true;
        }
        else if(hostname === "FOISX-PC" && (adDomain === "TREGCC" || adDomain === "tregcc.local")) {
            return true;
        }
        else {
            return false;
        }
    } else {
        return false;
    }
}

export const evaluateHostname = (hostname: string): boolean => {
    if (hostnameWhitelist.includes(hostname)) {
        return true;
    }

    return hostnameRegex.test(hostname);
}

export const isInsideADDomain = (adDomain: string): boolean => {
    return adDomain !== "WORKGROUP";
}

export const isInsideTREGCCADDomain = (adDomain: string): boolean => {
    if (!isInsideADDomain(adDomain)) {
        return false;
    }
    return adDomain === "TREGCC" || adDomain === "tregcc.local";
}

export type HostIntegrityStanding = "perfect" | "acceptable" | "limited";

// Single source of truth for the tri-state host integrity standing, shared
// between the automatic app-wide check (hostIntegrityCheck.ts) and the
// manual Safety Check dialog, so both always agree on the same result for
// the same inputs.
export const deriveHostIntegrityStanding = (params: {
    hostnameOk: boolean;
    adDomainOk: boolean;
    hostIntegrityToggleOk: boolean;
    updaterInstalled: boolean;
    updaterRunning: boolean;
    ipcActive: boolean;
    ipcResponding: boolean;
    crossCheckOk: boolean;
}): HostIntegrityStanding => {
    const {
        hostnameOk,
        adDomainOk,
        hostIntegrityToggleOk,
        updaterInstalled,
        updaterRunning,
        ipcActive,
        ipcResponding,
        crossCheckOk,
    } = params;

    if (!hostIntegrityToggleOk || !hostnameOk || !adDomainOk) {
        return "limited";
    }
    if (ipcActive && ipcResponding && !crossCheckOk) {
        return "limited";
    }
    if (updaterInstalled && updaterRunning && ipcActive && ipcResponding && crossCheckOk) {
        return "perfect";
    }
    return "acceptable";
}