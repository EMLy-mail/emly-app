export const hostnameRegex = /^(PCRM|RM)\d+$/;

export const hostnameWhitelist = [
    "FOISX-PC",
    "STRAN1-PC",
    "COSE1-PC",
    "RM094"
]

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
    return adDomain === "TREGCC";
}