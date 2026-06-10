import { error } from "@sveltejs/kit";

export const load = () => {
    throw error(500, "Test Internal Server Error");
};
