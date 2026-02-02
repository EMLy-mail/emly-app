<script lang="ts">
  import SettingsIcon from "@lucide/svelte/icons/settings";
  import * as Sidebar from "$lib/components/ui/sidebar/index.js";
  import { dangerZoneEnabled } from "$lib/stores/app";
  import * as m from "$lib/paraglide/messages.js";
  import { Mail } from "@lucide/svelte/icons";

  const CLICK_WINDOW_MS = 4000;
  const REQUIRED_CLICKS = 10;

  let recentClicks: number[] = [];

  function enableDangerZone(_event: MouseEvent) {
    const now = Date.now();

    recentClicks = recentClicks.filter((t) => now - t < CLICK_WINDOW_MS);
    recentClicks.push(now);

    if (recentClicks.length >= REQUIRED_CLICKS) {
      recentClicks = [];
      try {
        sessionStorage.setItem("debugWindowInSettings", "true");
        dangerZoneEnabled.set(true);
      } catch (e) {
        console.error("Failed to enable debug window:", e);
      }
    }
  }

  // Menu items.
  const items = [
    {
      title: m.sidebar_overview(),
      url: "/",
      icon: Mail,
      disabled: false,
      id: 1,
    },
    {
      title: m.sidebar_settings(),
      url: "/settings",
      icon: SettingsIcon,
      disabled: false,
      id: 2,
    },
  ];
</script>

<Sidebar.Root style="opacity: 0.8;">
  <Sidebar.Header>
    <div
      class="sidebar-title items-center justify-center p-3 border-b border-white/10"
      style="padding: 12px; border-bottom: 1px solid rgba(255, 255, 255, 0.08); display: flex; justify-content: center;"
    >
      <img src="/appicon.png" alt="Logo" width="64" height="64" />
      <span
        class="font-bold text-lg mt-2 pl-3"
        style="font-family: system-ui, sans-serif;">EMLy by 3gIT</span
      >
    </div>
  </Sidebar.Header>
  <Sidebar.Content>
    <Sidebar.Group>
      <Sidebar.GroupLabel>Menu</Sidebar.GroupLabel>
      <Sidebar.GroupContent>
        <Sidebar.Menu>
          {#each items as item (item.id)}
            <Sidebar.MenuItem>
              <Sidebar.MenuButton>
                {#snippet child({ props })}
                  {#if item.disabled}
                    <a aria-disabled={item.disabled} tabindex="-1" {...props}>
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  {:else if item.url === "/settings"}
                    <a href={item.url} {...props} onclick={enableDangerZone}>
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  {:else}
                    <a href={item.url} {...props}>
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  {/if}
                {/snippet}
              </Sidebar.MenuButton>
            </Sidebar.MenuItem>
          {/each}
        </Sidebar.Menu>
      </Sidebar.GroupContent>
    </Sidebar.Group>
  </Sidebar.Content>
</Sidebar.Root>
