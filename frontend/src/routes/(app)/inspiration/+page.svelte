<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Separator } from "$lib/components/ui/separator";
  import { ChevronLeft, Music, ExternalLink } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";
  import { OpenURLInBrowser } from "$lib/wailsjs/go/main/App";
  import type { SpotifyTrack } from "./+page";

  let { data } = $props();
  let config = $derived(data.config);
  let tracks: SpotifyTrack[] = $derived(data.tracks ?? []);

  // Open external URL in default browser
  async function openUrl(url: string) {
    try {
      await OpenURLInBrowser(url);
    } catch (e) {
      console.error("Failed to open URL:", e);
    }
  }
</script>

<div class="min-h-[calc(100vh-1rem)] bg-gradient-to-b from-background to-muted/30">
  <div
    class="mx-auto flex max-w-4xl flex-col gap-4 px-4 py-6 sm:px-6 sm:py-10 opacity-80"
  >
    <header class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <h1
          class="text-balance text-2xl font-semibold tracking-tight sm:text-3xl"
        >
          Musical Inspiration
        </h1>
        <p class="mt-2 text-sm text-muted-foreground">
          This project was mainly coded to the following tracks
        </p>
      </div>
      <Button
        class="cursor-pointer hover:cursor-pointer"
        variant="ghost"
        onclick={() => goto("/")}
      >
        <ChevronLeft class="size-4" /> Back
      </Button>
    </header>

    <Separator class="my-2" />

    <!-- Spotify Embeds -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="flex items-center gap-2">
          <Music class="size-5" />
          FOISX's Soundtrack
        </Card.Title>
        <Card.Description>
          The albums and tracks that fueled the development of EMLy
        </Card.Description>
      </Card.Header>
      <Card.Content>
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-2">
          {#each tracks as track}
            <div class="group relative">
              <div class="overflow-hidden rounded-lg bg-muted">
                {#if track.embedHtml}
                  {@html track.embedHtml}
                {:else}
                  <iframe
                    src={track.embedUrl}
                    width="100%"
                    height="352"
                    frameborder="0"
                    allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture"
                    loading="lazy"
                    title={`${track.artist} - ${track.name}`}
                    class="rounded-lg"
                  ></iframe>
                {/if}
              </div>
              
            </div>
          {/each}
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Fun fact -->
    <Card.Root class="border-primary/20 bg-primary/5">
      <Card.Content class="">
        <div class="flex items-start gap-3">
          <Music class="size-5 text-primary mt-0.5 shrink-0" />
          <div class="space-y-1">
            <p class="text-sm font-medium">The Soundtrack</p>
            <p class="text-sm text-muted-foreground">
              These are just a small sample of what helped inspire the project.
              Although they represent a wide variety of emotions, themes and genres, some exploring deep meanings
              of betrayal, personal struggles, and introspection, they provided solace and strength to the main developer
              during challenging times. 
              <br/>
              Music has a unique way of transforming pain into creative energy..
            </p>
          </div>
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Footer note -->
    <div class="text-center text-xs text-muted-foreground">
      <p>
        Made with
        <Music class="inline-block size-3 mx-1" />
        and
        <span class="text-red-500">♥</span>
      </p>
    </div>
  </div>
</div>

<style>
  iframe {
    border-radius: 0.5rem;
  }
</style>
