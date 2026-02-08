<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Separator } from "$lib/components/ui/separator";
  import { ChevronLeft, Music, ExternalLink } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";
  import { OpenURLInBrowser } from "$lib/wailsjs/go/main/App";

  let { data } = $props();
  let config = $derived(data.config);

  interface SpotifyTrack {
    name: string;
    artist: string;
    albumArt?: string;
    spotifyUrl: string;
    embedUrl: string;
  }

  // Music that inspired this project
  const inspirationTracks: SpotifyTrack[] = [
    {
      name: "Strays",
      artist: "Ivycomb, Stephanafro",
      spotifyUrl: "https://open.spotify.com/track/1aXATIo34e5ZZvFcavePpy",
      embedUrl: "https://open.spotify.com/embed/track/1aXATIo34e5ZZvFcavePpy?utm_source=generator"
    },
    {
      name: "Headlock",
      artist: "Imogen Heap",
      spotifyUrl: "https://open.spotify.com/track/63Pi2NAx5yCgeLhCTOrEou",
      embedUrl: "https://open.spotify.com/embed/track/63Pi2NAx5yCgeLhCTOrEou?utm_source=generator"
    },
    {
      name: "I Still Create",
      artist: "YonKaGor",
      spotifyUrl: "https://open.spotify.com/track/0IqTgwWU2syiSYbdBEromt",
      embedUrl: "https://open.spotify.com/embed/track/0IqTgwWU2syiSYbdBEromt?utm_source=generator"
    },
    {
      name: "Raised by Aliens",
      artist: "ivy comb, Stephanafro",
      spotifyUrl: "https://open.spotify.com/track/5ezyCaoc5XiVdkpRYWeyG5",
      embedUrl: "https://open.spotify.com/embed/track/5ezyCaoc5XiVdkpRYWeyG5?utm_source=generator"
    },
    {
      name: "VENOMOUS",
      artist: "passengerprincess",
      spotifyUrl: "https://open.spotify.com/track/4rPKifkzrhIYAsl1njwmjd",
      embedUrl: "https://open.spotify.com/embed/track/4rPKifkzrhIYAsl1njwmjd?utm_source=generator"
    },
    {
      name: "PREY",
      artist: "passengerprincess",
      spotifyUrl: "https://open.spotify.com/track/510m8qwFCHgzi4zsQnjLUX",
      embedUrl: "https://open.spotify.com/embed/track/510m8qwFCHgzi4zsQnjLUX?utm_source=generator"
    },
    {
      name: "Dracula",
      artist: "Tame Impala",
      spotifyUrl: "https://open.spotify.com/track/1NXbNEAcPvY5G1xvfN57aA",
      embedUrl: "https://open.spotify.com/embed/track/1NXbNEAcPvY5G1xvfN57aA?utm_source=generator"
    },
    {
      name: "Electric love",
      artist: "When Snakes Sing",
      spotifyUrl: "https://open.spotify.com/track/1nDkT2Cn13qDnFegF93UHi",
      embedUrl: "https://open.spotify.com/embed/track/1nDkT2Cn13qDnFegF93UHi?utm_source=generator"
    }
  ];

  // Open external URL in default browser
  async function openUrl(url: string) {
    try {
      await OpenURLInBrowser(url);
    } catch (e) {
      console.error("Failed to open URL:", e);
    }
  }
</script>

<div class="min-h-[calc(100vh-1rem)] from-background to-muted/30">
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
          {#each inspirationTracks as track}
            <div class="group relative">
              <div class="overflow-hidden rounded-lg bg-muted" style="height: 352px;">
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
              </div>
              <div class="mt-2 flex items-start justify-between gap-2">
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm font-medium">{track.name}</p>
                  <p class="truncate text-xs text-muted-foreground">
                    {track.artist}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-8 shrink-0 opacity-70 hover:opacity-100"
                  onclick={() => openUrl(track.spotifyUrl)}
                  title="Open in Spotify"
                >
                  <ExternalLink class="size-4" />
                </Button>
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
      <p class="mt-1">
        GUI: {config ? `v${config.GUISemver}` : "N/A"} • 
        SDK: {config ? `v${config.SDKDecoderSemver}` : "N/A"}
      </p>
    </div>
  </div>
</div>

<style>
  iframe {
    border-radius: 0.5rem;
  }
</style>
