import type { PageLoad } from './$types';
import { GetConfig } from "$lib/wailsjs/go/main/App";
import { browser } from '$app/environment';
import type { utils } from '$lib/wailsjs/go/models';

interface GitHubUserAssignment {
    id: number;
    role: "team" | "specialThanks";
}

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

interface ContributorsData {
    team: Array<GitHubUserData | null>;
    specialThanks: Array<GitHubUserData | null>;
}

export const load = (async ({fetch}) => {
    if (!browser) return {
        config: null, contributorsData: {
            team: [], specialThanks: []
        }
    };

    const ghUserId: GitHubUserAssignment[] = [{
        id: 278996585, // LyzCoote
        role: "team"
    }, {
        id: 35636667, // Lauren
        role: "specialThanks"
    }, {
        id: 20886839, // Amber
        role: "specialThanks"
    }]

    let configRoot: utils.Config | null = null;

    const contributorsData: ContributorsData = {
        team: [],
        specialThanks: []
    };

    async function fetchGitHubUserData(ghUserId: GitHubUserAssignment): Promise<GitHubUserData | null> {
        return new Promise<GitHubUserData | null>(async (resolve) => {
            try {
                const response = await fetch(`https://api.github.com/user/${ghUserId.id}`, {
                    headers: {
                        "Accept": "application/vnd.github+json",
                        "X-GitHub-Api-Version": "2026-03-10"
                    }
                });
                if (response.ok) {
                    const data = await response.json();
                    resolve(data);
                } else {
                    console.error(`Failed to fetch GitHub profile for user ID ${ghUserId.id}: HTTP ${response.status}`);
                    resolve(null);
                }
            } catch (e) {
                console.error(`Failed to fetch GitHub profile for user ID ${ghUserId.id}:`, e);
                resolve(null);
            }
        });
    }

    async function processArray(array: GitHubUserAssignment[]): Promise<void> {
        for (const item of array) {
            let res: GitHubUserData | null = await fetchGitHubUserData(item);
            if(item.role === "team") {
                contributorsData.team.push(res);
            } else if(item.role === "specialThanks") {
                contributorsData.specialThanks.push(res);
            }
        }
    }

    try {
        configRoot = await GetConfig();
    } catch (e) {
        console.error("Failed to load config for credits", e);
    }

    try {
        await processArray(ghUserId);
    } catch (e) {
        console.error("Failed to load GitHub user data for credits", e);
    }

    return {
        config: configRoot?.EMLy,
        contributorsData
    };


}) satisfies PageLoad;
