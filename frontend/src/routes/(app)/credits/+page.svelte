<script lang="ts">
  import { goto, preloadData } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Separator } from "$lib/components/ui/separator";
  import { ChevronLeft, Heart, Code, Package, Globe, Github, Mail, BadgeInfo, Music, PartyPopper } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";
  import { OpenURLInBrowser } from "$lib/wailsjs/go/main/App";
  import { dangerZoneEnabled } from "$lib/stores/app";
  import { settingsStore } from "$lib/stores/settings.svelte";
  import { toast } from "svelte-sonner";

  let { data } = $props();
  let config = $derived(data.config);

  // Easter Egg State
  const REQUIRED_CLICKS = 10;
  const CLICK_WINDOW_MS = 4000;
  let recentClicks: number[] = [];

  function handleEasterEggClick(_event: MouseEvent) {
    console.log("clicked")
    // Only proceed if danger zone is already enabled
    if (!$dangerZoneEnabled) return;
    
    // If already enabled, do nothing to avoid spam
    if (settingsStore.settings.musicInspirationEnabled) return;

    const now = Date.now();
    
    // Clean old clicks
    recentClicks = recentClicks.filter(t => now - t < CLICK_WINDOW_MS);
    recentClicks.push(now);

    if (recentClicks.length >= REQUIRED_CLICKS) {
      recentClicks = [];
      try {
        settingsStore.update({ musicInspirationEnabled: true });
        preloadData("/inspiration");
      } catch (e) {
        console.error("Failed to enable music inspiration:", e);
      }
    }
  }

  // Open external URL in default browser
  async function openUrl(url: string) {
    try {
      await OpenURLInBrowser(url);
    } catch (e) {
      console.error("Failed to open URL:", e);
    }
  }

  // Gravatar URL helper - uses MD5 hash of email
  // Pre-computed hashes for known emails
  const gravatarUrls: Record<string, string> = {
    "f.fois@3git.eu": "https://gravatar.com/avatar/6a2b6cfd8ab2c36ac3eace1faa871f79084b64ad08fb6e490f050e71ee1b599c",
    "iraci.matteo@gmail.com": "https://gravatar.com/avatar/0c17334ae886eb44b670d226e7de32ac082b9c85925ce4ed4c12239d9d8351f2",
  };

  // Technology stack
  const technologies = [
    { name: "Wails v2", description: m.credits_tech_wails(), url: "https://wails.io" },
    { name: "Go", description: m.credits_tech_go(), url: "https://go.dev" },
    { name: "SvelteKit", description: m.credits_tech_sveltekit(), url: "https://kit.svelte.dev" },
    { name: "Svelte 5", description: m.credits_tech_svelte(), url: "https://svelte.dev" },
    { name: "TypeScript", description: m.credits_tech_typescript(), url: "https://www.typescriptlang.org" },
    { name: "Tailwind CSS", description: m.credits_tech_tailwind(), url: "https://tailwindcss.com" },
  ];

  // Libraries and packages
  const libraries = [
    { name: "shadcn-svelte", description: m.credits_lib_shadcn(), url: "https://www.shadcn-svelte.com" },
    { name: "Lucide Icons", description: m.credits_lib_lucide(), url: "https://lucide.dev" },
    { name: "ParaglideJS", description: m.credits_lib_paraglide(), url: "https://inlang.com/m/gerre34r/library-inlang-paraglideJs" },
    { name: "svelte-sonner", description: m.credits_lib_sonner(), url: "https://svelte-sonner.vercel.app" },
    { name: "PDF.js", description: m.credits_lib_pdfjs(), url: "https://mozilla.github.io/pdf.js" },
    { name: "DOMPurify", description: m.credits_lib_dompurify(), url: "https://github.com/cure53/DOMPurify" },
  ];

  // Team / Contributors
  const team = [
    {
      username: "FOISX",
      name: "Flavio Fois",
      role: m.credits_role_lead_developer(),
      description: m.credits_foisx_desc(),
      email: "f.fois@3git.eu",
    },
  ];

  // Special thanks
  const specialThanks = [
    {
      name: "Laky64",
      contribution: m.credits_laky64_desc(),
      email: "iraci.matteo@gmail.com",
    },
  ];
</script>

<div class="min-h-[calc(100vh-1rem)] bg-gradient-to-b from-background to-muted/30">
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
          <span style="font-style: italic">{m.credits_about_description()}</span>
          <span>{m.credits_about_description_2()}</span>
        </Card.Description>
      </Card.Header>
      <Card.Content>
        <div class="flex items-center gap-4 mb-4">
          <img src="/appicon.png" alt="EMLy Logo" width="64" height="64" class="rounded-lg" />
          <div>
            <h3 class="font-semibold text-lg">EMLy</h3>
            <p class="text-sm text-muted-foreground">{m.credits_app_tagline()}</p>
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
    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title class="flex items-center gap-2">
          <Code class="size-5" />
          {m.credits_team_title()}
        </Card.Title>
        <Card.Description>{m.credits_team_description()}</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        {#each team as member}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div 
            class="flex items-start gap-4 rounded-lg border bg-card p-4 relative overflow-hidden"
            onclick={member.username === "FOISX" ? handleEasterEggClick : undefined}
          >
            <!-- Selectable trigger area overlay for cleaner interaction -->
            {#if member.username === "FOISX" && $dangerZoneEnabled && !settingsStore.settings.musicInspirationEnabled}
               <div class="absolute inset-0 cursor-pointer z-10 opacity-0 bg-transparent"></div>
            {/if}
            
            <img
              src={gravatarUrls[member.email]}
              alt={member.name}
              class="h-14 w-14 rounded-full border-2 border-primary/20 z-0 select-none"
            />
            <div class="flex-1 z-0">
              <div class="font-medium">{member.username} ({member.name})</div>
              <div class="text-sm text-primary/80">{member.role}</div>
              <div class="text-sm text-muted-foreground mt-1">{member.description}</div>
              <a
                href="mailto:{member.email}"
                class="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-primary mt-2 transition-colors relative z-20"
              >
                <Mail class="size-3" />
                {member.email}
              </a>
            </div>
          </div>
        {/each}
        <div class="text-center text-sm text-muted-foreground pt-2">
          <span class="flex items-center justify-center gap-1">
            {m.credits_made_with()} <Heart class="size-3 text-red-500 inline" /> {m.credits_at_3git()}
          </span>
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Special Thanks Card -->
    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title class="flex items-center gap-2">
          <Heart class="size-5 text-pink-500" />
          {m.credits_special_thanks_title()}
        </Card.Title>
        <Card.Description>{m.credits_special_thanks_description()}</Card.Description>
      </Card.Header>
      <Card.Content>
        <div class="space-y-3">
          {#each specialThanks as contributor}
            <div class="flex items-center gap-3 rounded-lg border bg-card p-3">
              <img
                src={gravatarUrls[contributor.email]}
                alt={contributor.name}
                class="h-10 w-10 rounded-full border-2 border-primary/20"
              />
              <div class="flex-1">
                <span class="font-medium text-sm">{contributor.name}</span>
                -
                <span class="text-muted-foreground text-sm">{contributor.contribution}</span>
              </div>
            </div>
          {/each}
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Technologies Card -->
    <Card.Root>
      <Card.Header class="space-y-1">
        <Card.Title class="flex items-center gap-2">
          <Globe class="size-5" />
          {m.credits_tech_title()}
        </Card.Title>
        <Card.Description>{m.credits_tech_description()}</Card.Description>
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
                <div class="font-medium text-sm">{tech.name}</div>
                <div class="text-xs text-muted-foreground">{tech.description}</div>
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
        <Card.Description>{m.credits_libraries_description()}</Card.Description>
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
                <div class="font-medium text-sm">{lib.name}</div>
                <div class="text-xs text-muted-foreground">{lib.description}</div>
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
          <Github class="size-5" />
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
