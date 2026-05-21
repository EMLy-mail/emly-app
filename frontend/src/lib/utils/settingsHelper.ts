import * as m from "$lib/paraglide/messages";

export const UNC_PATH_RE = /^\\\\[^\\/]+\\[^\\/]+/;
export const LOCAL_PATH_RE = /^[A-Za-z]:\\/;

export const UPDATE_PATH_OPTIONS = {
    "DC-RM2": "\\\\dc-rm2\\logo\\update",
    "DC-CB": "\\\\dc-cb\\logo\\update",
    Other: "",
} as const;
export type UpdatePathOption = keyof typeof UPDATE_PATH_OPTIONS;
export const UPDATE_PATH_LABELS: Record<UpdatePathOption, string> = {
    "DC-RM2": `DC-RM2 (${UPDATE_PATH_OPTIONS["DC-RM2"]})`,
    "DC-CB": `DC-CB (${UPDATE_PATH_OPTIONS["DC-CB"]})`,
    Other: m.settings_update_select_other(),
};