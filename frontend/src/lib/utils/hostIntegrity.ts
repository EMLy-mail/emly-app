export const hostnameRegex = /^(PCRM|RM)\d+$/;

export const hostnameWhitelist = [
    "STRAN1-PC",
    "COSE1-PC",
    "RM094",
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
        else if(hostname === "FOISX-PC" && adDomain === "TREGCC") {
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
    return adDomain === "TREGCC";
}