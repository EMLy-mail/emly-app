import * as m from "$lib/paraglide/messages";

export interface GitHubUserData {
    login: string;
    id: number;
    node_id: string;
    avatar_url: string;
    gravatar_id: string;
    url: string;
    html_url: string;
    followers_url: string;
    following_url: string;
    gists_url: string;
    starred_url: string;
    subscriptions_url: string;
    organizations_url: string;
    repos_url: string;
    events_url: string;
    received_events_url: string;
    type: "User" | "Organization" | "Bot";
    user_view_type: "public" | "private";
    site_admin: boolean;
    name: string | null;
    company: string | null;
    blog: string | null;
    location: string | null;
    email: string | null;
    hireable: boolean | null;
    bio: string | null;
    twitter_username: string | null;
    public_repos: number;
    public_gists: number;
    followers: number;
    following: number;
    created_at: string;
    updated_at: string;
}

export interface ContributorAssignment {
    id: number;
    role: "team" | "specialThanks";
}

export interface ContributorProfile {
    role: "team" | "specialThanks";
    roleLabel: () => string;
    description: () => string;
    /**
     * Overrides the GitHub-reported email on the credits page. Only
     * LyzCoote uses this today (the profile shows f.fois@3git.eu instead
     * of whatever email the GitHub API returns for that account) -
     * preserved as-is from before this was unified, not a new behavior.
     */
    emailOverride?: string;
}

/**
 * Single source of truth for the three GitHub accounts shown on the
 * credits page: which GitHub user ID to fetch, which section it belongs
 * in, and what to render alongside it. Previously this was split across
 * a literal assignment array and duplicated `id === ...` checks in
 * +page.svelte.
 */
export const CONTRIBUTOR_PROFILES: Record<number, ContributorProfile> = {
    278996585: {
        // LyzCoote
        role: "team",
        roleLabel: () => m.credits_role_lead_developer(),
        description: () => m.credits_foisx_desc(),
        emailOverride: "f.fois@3git.eu",
    },
    35636667: {
        // Lauren
        role: "specialThanks",
        roleLabel: () => m.credits_role_go_contributor(),
        description: () => m.credits_laky64_desc(),
    },
    20886839: {
        // Amber
        role: "specialThanks",
        roleLabel: () => m.credits_role_ui_ux_feedback_advisor(),
        description: () => m.credits_amber_desc(),
    },
};

export const CONTRIBUTOR_ASSIGNMENTS: ContributorAssignment[] = Object.entries(
    CONTRIBUTOR_PROFILES,
).map(([id, profile]) => ({ id: Number(id), role: profile.role }));
