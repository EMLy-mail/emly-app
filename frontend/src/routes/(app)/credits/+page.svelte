<script lang="ts">
    import { goto } from "$app/navigation";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { Separator } from "$lib/components/ui/separator";
    import {
        ChevronLeft,
        Heart,
        Code,
        Package,
        Globe,
        GitMerge,
        BadgeInfo,
    } from "@lucide/svelte";
    import * as m from "$lib/paraglide/messages";
    import { OpenURLInBrowser } from "$lib/wailsjs/go/main/App";
    import "@risadams/pride-flags/dist/pride-flags.css";
    import ContributorCard from "$lib/components/ContributorCard.svelte";
    import { CONTRIBUTOR_PROFILES } from "$lib/data/contributors";

    let { data } = $props();
    let config = $derived(data.config);

    let easterEgg = $derived(data.easterEgg ?? false);

    // Open external URL in default browser
    async function openUrl(url: string) {
        let sanitizedUrl: URL = new URL(url);
        if (!["http:", "https:"].includes(sanitizedUrl.protocol)) {
            console.warn(
                "Attempted to open URL with unsupported protocol:",
                url,
            );
            return;
        }
        const urlPattern = /^[a-zA-Z0-9-._~:/?#[\]@!$&'()*+,;=]+$/;
        if (!urlPattern.test(sanitizedUrl.href)) {
            console.warn(
                "Attempted to open URL with potentially unsafe characters:",
                url,
            );
            return;
        }
        try {
            await OpenURLInBrowser(sanitizedUrl.href);
        } catch (e) {
            console.error("Failed to open URL:", e);
        }
    }

    let teamData = $derived(
        (data.contributorsData.team ?? []).filter(
            (member): member is NonNullable<typeof member> => !!member,
        ),
    );
    let specialThanksData = $derived(
        (data.contributorsData.specialThanks ?? []).filter(
            (contributor): contributor is NonNullable<typeof contributor> =>
                !!contributor,
        ),
    );

    let gravatarUrls: Record<string, string> = $derived.by(() => {
        const urls: Record<string, string> = {};
        for (const [role, contributors] of [
            ["team", teamData],
            ["specialThanks", specialThanksData],
        ] as const) {
            for (const contributor of contributors) {
                if (!contributor) continue;
                urls[contributor.id] = contributor.avatar_url;
            }
        }
        return urls;
    });

    const technologies = [
        {
            name: "Wails v2",
            description: m.credits_tech_wails(),
            url: "https://wails.io",
        },
        { name: "Go", description: m.credits_tech_go(), url: "https://go.dev" },
        {
            name: "SvelteKit",
            description: m.credits_tech_sveltekit(),
            url: "https://kit.svelte.dev",
        },
        {
            name: "Svelte 5",
            description: m.credits_tech_svelte(),
            url: "https://svelte.dev",
        },
        {
            name: "TypeScript",
            description: m.credits_tech_typescript(),
            url: "https://www.typescriptlang.org",
        },
        {
            name: "Tailwind CSS",
            description: m.credits_tech_tailwind(),
            url: "https://tailwindcss.com",
        },
    ];

    const libraries = [
        {
            name: "shadcn-svelte",
            description: m.credits_lib_shadcn(),
            url: "https://www.shadcn-svelte.com",
        },
        {
            name: "Lucide Icons",
            description: m.credits_lib_lucide(),
            url: "https://lucide.dev",
        },
        {
            name: "ParaglideJS",
            description: m.credits_lib_paraglide(),
            url: "https://inlang.com/m/gerre34r/library-inlang-paraglideJs",
        },
        {
            name: "svelte-sonner",
            description: m.credits_lib_sonner(),
            url: "https://svelte-sonner.vercel.app",
        },
        {
            name: "PDF.js",
            description: m.credits_lib_pdfjs(),
            url: "https://github.com/mozilla/pdf.js",
        },
        {
            name: "DOMPurify",
            description: m.credits_lib_dompurify(),
            url: "https://github.com/cure53/DOMPurify",
        },
    ];
</script>

<div class="min-h-[calc(100vh-1rem)] from-background to-muted/30">
    <div
        class="mx-auto flex max-w-3xl flex-col gap-4 px-4 py-6 sm:px-6 sm:py-10 opacity-80"
    >
        <header class="flex items-start justify-between gap-3">
            <div class="min-w-0">
                <h1
                    class="text-balance text-2xl font-semibold tracking-tight sm:text-3xl"
                >
                    {m.credits_title()}
                </h1>
                <p class="mt-2 text-sm text-muted-foreground">
                    {m.credits_description()}
                </p>
            </div>
            <Button
                class="cursor-pointer hover:cursor-pointer"
                variant="ghost"
                onclick={() => goto("/")}
                ><ChevronLeft class="size-4" /> {m.settings_back()}</Button
            >
        </header>

        <!-- About Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <BadgeInfo class="size-5" />
                    {m.credits_about_title()}
                </Card.Title>
                <Card.Description>
                    <span style="font-style: italic"
                        >{m.credits_about_description()}</span
                    >
                    <span>{m.credits_about_description_2()}</span>
                </Card.Description>
            </Card.Header>
            <Card.Content>
                <div class="flex items-center gap-4 mb-4">
                    <img
                        src="/appicon.png"
                        alt={m.credits_logo_alt()}
                        width="64"
                        height="64"
                        class="rounded-lg"
                    />
                    <div>
                        <h3 class="font-semibold text-lg">EMLy</h3>
                        <p class="text-sm text-muted-foreground">
                            {m.credits_app_tagline()}
                        </p>
                        {#if config}
                            <p class="text-xs text-muted-foreground mt-1">
                                v{config.GUISemver} ({config.GUIReleaseChannel})
                            </p>
                        {/if}
                    </div>
                </div>
                <p class="text-sm text-muted-foreground">
                    {m.credits_app_description()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Team Card -->
        {#if teamData.length > 0}
            <Card.Root>
                <Card.Header class="space-y-1">
                    <Card.Title class="flex items-center gap-2">
                        <Code class="size-5" />
                        {m.credits_team_title()}
                    </Card.Title>
                    <Card.Description
                        >{m.credits_team_description()}</Card.Description
                    >
                </Card.Header>
                <Card.Content class="space-y-4">
                    {#each teamData as member}
                        {#if member}
                            <ContributorCard
                                contributor={member}
                                avatarUrl={gravatarUrls[member.id]}
                                profile={CONTRIBUTOR_PROFILES[member.id]}
                                showLogin
                            />
                        {/if}
                    {/each}
                    <div class="text-center text-sm text-muted-foreground pt-2">
                        <span class="flex items-center justify-center gap-1">
                            {m.credits_made_with()}
                            {#if easterEgg}
                                <Heart class="size-3 text-red-500 inline" />
                                by a
                                <div
                                    class="flag icon nonbinary"
                                    style="height: 0.75rem; width: auto;"
                                    role="img"
                                    aria-label="Non-Binary Pride Flag"
                                ></div>
                                <div
                                    class="flag icon pansexual"
                                    style="height: 0.75rem; width: auto;"
                                    role="img"
                                    aria-label="Pansexual Pride Flag"
                                ></div>
                                Protogen :3
                            {:else}
                                <Heart class="size-3 text-red-500 inline" />
                                {m.credits_at_3git()}
                            {/if}
                        </span>
                    </div>
                </Card.Content>
            </Card.Root>
        {/if}

        {#if specialThanksData.length > 0}
            <!-- Special Thanks Card -->
            <Card.Root>
                <Card.Header class="space-y-1">
                    <Card.Title class="flex items-center gap-2">
                        <Heart class="size-5 text-pink-500" />
                        {m.credits_special_thanks_title()}
                    </Card.Title>
                    <Card.Description>
                        {#if easterEgg}
                            {m.credits_special_thanks_description_alt()}
                        {:else}
                            {m.credits_special_thanks_description()}
                        {/if}
                    </Card.Description>
                </Card.Header>
                <Card.Content>
                    <div class="space-y-3">
                        {#each specialThanksData as contributor}
                            {#if contributor}
                                <ContributorCard
                                    {contributor}
                                    avatarUrl={gravatarUrls[contributor.id]}
                                    profile={CONTRIBUTOR_PROFILES[
                                        contributor.id
                                    ]}
                                />
                            {/if}
                        {/each}
                    </div>
                </Card.Content>
            </Card.Root>
        {/if}

        <!-- Technologies Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <Globe class="size-5" />
                    {m.credits_tech_title()}
                </Card.Title>
                <Card.Description
                    >{m.credits_tech_description()}</Card.Description
                >
            </Card.Header>
            <Card.Content>
                <div class="grid gap-3 sm:grid-cols-2">
                    {#each technologies as tech}
                        <button
                            type="button"
                            onclick={() => openUrl(tech.url)}
                            class="flex items-start gap-3 rounded-lg border bg-card p-3 transition-colors hover:bg-accent/50 cursor-pointer text-left"
                        >
                            <div class="flex-1">
                                <div class="font-medium text-sm">
                                    {tech.name}
                                </div>
                                <div class="text-xs text-muted-foreground">
                                    {tech.description}
                                </div>
                            </div>
                        </button>
                    {/each}
                </div>
            </Card.Content>
        </Card.Root>

        <!-- Libraries Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <Package class="size-5" />
                    {m.credits_libraries_title()}
                </Card.Title>
                <Card.Description
                    >{m.credits_libraries_description()}</Card.Description
                >
            </Card.Header>
            <Card.Content>
                <div class="grid gap-3 sm:grid-cols-2">
                    {#each libraries as lib}
                        <button
                            type="button"
                            onclick={() => openUrl(lib.url)}
                            class="flex items-start gap-3 rounded-lg border bg-card p-3 transition-colors hover:bg-accent/50 cursor-pointer text-left"
                        >
                            <div class="flex-1">
                                <div class="font-medium text-sm">
                                    {lib.name}
                                </div>
                                <div class="text-xs text-muted-foreground">
                                    {lib.description}
                                </div>
                            </div>
                        </button>
                    {/each}
                </div>
            </Card.Content>
        </Card.Root>

        <!-- License Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <GitMerge class="size-5" />
                    {m.credits_license_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.credits_license_text()}
                </p>
                <Separator class="my-4" />
                <p class="text-xs text-muted-foreground text-center">
                    © 2025-{new Date().getFullYear()} 3gIT. {m.credits_copyright()}
                </p>
            </Card.Content>
        </Card.Root>
    </div>
</div>
