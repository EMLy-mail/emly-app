<script lang="ts">
    import { Mail } from "@lucide/svelte";
    import type {
        GitHubUserData,
        ContributorProfile,
    } from "$lib/data/contributors";

    let {
        contributor,
        avatarUrl,
        profile,
        showLogin = false,
    }: {
        contributor: GitHubUserData;
        avatarUrl: string;
        profile: ContributorProfile | undefined;
        showLogin?: boolean;
    } = $props();

    let email = $derived(profile?.emailOverride ?? contributor.email);
</script>

<div
    class="flex items-start gap-4 rounded-lg border bg-card p-4 relative overflow-hidden"
>
    <img
        src={avatarUrl}
        alt={contributor.name}
        class="h-14 w-14 rounded-full border-2 border-primary/20 z-0 select-none"
    />
    <div class="flex-1 z-0">
        <div class="font-medium">
            {contributor.name}{showLogin ? ` (${contributor.login})` : ""}
        </div>
        {#if profile}
            <div class="text-sm text-primary/80">{profile.roleLabel()}</div>
            <div class="text-sm text-muted-foreground mt-1">
                {profile.description()}
            </div>
            {#if email}
                <a
                    href="mailto:{email}"
                    class="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-primary mt-2 transition-colors relative z-20"
                >
                    <Mail class="size-3" />
                    {email}
                </a>
            {/if}
        {/if}
    </div>
</div>
