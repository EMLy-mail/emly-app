import type { PageLoad } from './$types';
import { GetConfig } from "$lib/wailsjs/go/main/App";
import { browser } from '$app/environment';
import type { utils } from '$lib/wailsjs/go/models';
import {
    CONTRIBUTOR_ASSIGNMENTS,
    type ContributorAssignment,
    type GitHubUserData,
} from '$lib/data/contributors';

interface ContributorsData {
    team: Array<GitHubUserData | null>;
    specialThanks: Array<GitHubUserData | null>;
}

export const load = (async ({ fetch }) => {
    if (!browser) return {
        config: null, contributorsData: {
            team: [], specialThanks: []
        }
    };

    let configRoot: utils.Config | null = null;

    const contributorsData: ContributorsData = {
        team: [],
        specialThanks: []
    };

    async function fetchGitHubUserData(assignment: ContributorAssignment): Promise<GitHubUserData | null> {
        return new Promise<GitHubUserData | null>(async (resolve) => {
            try {
                const response = await fetch(`https://api.github.com/user/${assignment.id}`, {
                    headers: {
                        "Accept": "application/vnd.github+json",
                        "X-GitHub-Api-Version": "2026-03-10"
                    }
                });
                if (response.ok) {
                    const data = await response.json();
                    resolve(data);
                } else {
                    console.error(`Failed to fetch GitHub profile for user ID ${assignment.id}: HTTP ${response.status}`);
                    resolve(null);
                }
            } catch (e) {
                console.error(`Failed to fetch GitHub profile for user ID ${assignment.id}:`, e);
                resolve(null);
            }
        });
    }

    async function processArray(array: ContributorAssignment[]): Promise<void> {
        for (const item of array) {
            let res: GitHubUserData | null = await fetchGitHubUserData(item);
            if (item.role === "team") {
                contributorsData.team.push(res);
            } else if (item.role === "specialThanks") {
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
        await processArray(CONTRIBUTOR_ASSIGNMENTS);
    } catch (e) {
        console.error("Failed to load GitHub user data for credits", e);
    }

    let juneTriggerFloat = 0.2; // 20% chance of triggering the easter egg in June
    let normalTriggerFloat = 0.05; // 5% chance of triggering the easter egg in other months

    function getTriggerFloat(): number {
        const currentMonth = new Date().getMonth() + 1;
        if (currentMonth === 6) {
            return juneTriggerFloat;
        } else {
            return normalTriggerFloat;
        }
    }

    let easterEgg = false;

    if (Math.random() < getTriggerFloat()) {
        easterEgg = true;

        if (contributorsData.team[0] && contributorsData.team[0].id === 278996585) {
            contributorsData.team[0].avatar_url = "https://avatars.githubusercontent.com/u/44366896?v=4&s=400&u=1c9e5b8a7c3d2e5f8b6a9c4d2e1f3a4b5c6d7e&v=4";
        }

        for (let i = 0; i < contributorsData.specialThanks.length; i++) {
            if (contributorsData.specialThanks[i]) {
                contributorsData.specialThanks[i]!.name = `${contributorsData.specialThanks[i]!.name} (meow)`;
            }
        }
    }

    return {
        config: configRoot?.EMLy,
        contributorsData,
        easterEgg
    };


}) satisfies PageLoad;
