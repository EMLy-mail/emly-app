import type { PageLoad } from './$types';
import { GetConfig } from "$lib/wailsjs/go/main/App";
import { browser } from '$app/environment';

export interface SpotifyTrack {
    name: string;
    artist: string;
    spotifyUrl: string;
    embedUrl: string;
    embedHtml?: string;
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

async function fetchEmbedHtml(track: SpotifyTrack, fetch: typeof globalThis.fetch): Promise<SpotifyTrack> {
    try {
        const oEmbedUrl = `https://open.spotify.com/oembed?url=${encodeURIComponent(track.spotifyUrl)}`;
        const res = await fetch(oEmbedUrl);
        if (res.ok) {
            const data = await res.json();
            return { ...track, embedHtml: data.html };
        }
    } catch (e) {
        console.error(`Failed to fetch oEmbed for ${track.name}:`, e);
    }
    return track;
}

export const load = (async ({fetch}) => {
    if (!browser) return { config: null, tracks: inspirationTracks };

    try {
        const [configRoot, ...tracks] = await Promise.all([
            GetConfig(),
            ...inspirationTracks.map(t => fetchEmbedHtml(t, fetch))
        ]);

        return {
            config: configRoot.EMLy,
            tracks
        };
    } catch (e) {
        console.error("Failed to load data for inspiration", e);
        return {
            config: null,
            tracks: inspirationTracks
        };
    }
}) satisfies PageLoad;
