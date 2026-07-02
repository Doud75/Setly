<script lang="ts">
    import type { PageData } from './$types';
    import { formatDuration } from '$lib/utils/utils';
    import type {Song} from "$lib/types";
    import { enhance } from '$app/forms';

    let { data }: { data: PageData } = $props();

    let songs = $derived(data.songs ?? []);

    const songsByAlbum = $derived.by(() => {
        const grouped: Record<string, Song[]> = {};
        for (const song of songs) {
        const album = song.album_name ?? 'Sans Album';
            if (!grouped[album]) {
                grouped[album] = [];
            }
            grouped[album].push(song);
        }
        const sortedKeys = Object.keys(grouped).sort((a, b) => {
            if (a === 'Sans Album') return 1;
            if (b === 'Sans Album') return -1;
            return a.localeCompare(b);
        });
        const sortedGrouped: Record<string, Song[]> = {};
        for (const key of sortedKeys) {
            sortedGrouped[key] = grouped[key];
        }
        return sortedGrouped;
    });


</script>

<div class="container mx-auto px-4 sm:px-6">
    <header class="mb-8">
        <div class="flex flex-wrap items-center justify-between gap-4">
            <div>
                <h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
                    Bibliothèque de chansons
                </h1>
                <div class="mt-2 flex items-center gap-4 text-sm text-slate-500 dark:text-slate-400">
                    <a href="/" class="hover:underline">&larr; Retour à l'accueil</a>
                </div>
            </div>
            <a
                    href="/song/new"
                    class="flex w-auto justify-center rounded-md bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-indigo-500"
            >
                + Ajouter une chanson
            </a>
        </div>
    </header>

    <div class="space-y-8">
        {#if Object.keys(songsByAlbum).length > 0}
            {#each Object.entries(songsByAlbum) as [album, albumSongs] (album)}
                <div class="rounded-xl bg-white p-6 shadow-lg dark:bg-slate-800">
                    <h2 class="text-xl font-semibold text-slate-800 dark:text-slate-100">{album}</h2>
                    <ul class="mt-4 divide-y divide-slate-200 dark:divide-slate-700">
                        {#each albumSongs as song (song.id)}
                            <li class="flex items-center justify-between gap-3 py-3">
                                <div class="min-w-0 flex-grow">
                                    <div class="flex items-center gap-3">
                                        <a href="/song/{song.id}" class="truncate font-semibold text-indigo-600 hover:underline dark:text-indigo-400">{song.title}</a>
                                    </div>
                                    <div class="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-500 dark:text-slate-400">
                                        {#if song.duration_seconds !== null}
                                            <span>Durée: {formatDuration(song.duration_seconds)}</span>
                                        {/if}
                                        {#if song.tempo !== null}
                                            <span class="hidden sm:inline">&bull;</span>
                                            <span>Tempo: {song.tempo} BPM</span>
                                        {/if}
                                        {#if song.song_key !== null}
                                            <span class="hidden sm:inline">&bull;</span>
                                            <span>Tonalité: {song.song_key}</span>
                                        {/if}
                                        {#if song.links !== null}
                                            <span class="hidden sm:inline">&bull;</span>
                                            <a href={song.links} target="_blank" rel="noopener noreferrer" class="hover:underline">Lien</a>
                                        {/if}
                                    </div>
                                </div>
                                <div class="flex items-center gap-2">
                                    <a
                                            href="/song/{song.id}/edit"
                                            class="rounded-md p-2 text-slate-400 hover:bg-slate-100 hover:text-slate-600 dark:hover:bg-slate-700 dark:hover:text-slate-200"
                                            aria-label="Modifier {song.title}"
                                    >
                                        <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                fill="none"
                                                viewBox="0 0 24 24"
                                                stroke-width="1.5"
                                                stroke="currentColor"
                                                class="h-5 w-5"
                                        ><path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10"
                                        /></svg
                                        >
                                    </a>

                                    {#if data.user?.role === 'admin'}
                                    <form method="POST" action="?/deleteSong" use:enhance>
                                        <input type="hidden" name="songId" value={song.id} />
                                        <button
                                                type="submit"
                                                class="rounded-md p-2 text-slate-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10 dark:hover:text-red-400"
                                                aria-label="Supprimer {song.title}"
                                        >
                                            <svg
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    fill="none"
                                                    viewBox="0 0 24 24"
                                                    stroke-width="1.5"
                                                    stroke="currentColor"
                                                    class="h-5 w-5"
                                            ><path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.134-2.033-2.134H8.033C6.91 2.75 6 3.704 6 4.874v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
                                            /></svg
                                            >
                                        </button>
                                    </form>
                                    {/if}
                                </div>
                            </li>
                        {/each}
                    </ul>
                </div>
            {/each}
        {:else}
            <div
                    class="rounded-lg border-2 border-dashed border-slate-300 p-12 text-center dark:border-slate-700"
            >
                <svg
                        class="mx-auto h-12 w-12 text-slate-400"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        aria-hidden="true"
                >
                    <path
                            vector-effect="non-scaling-stroke"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="m9 9 10.5-3m0 6.553v3.75a2.25 2.25 0 0 1-1.632 2.163l-1.32.377a1.803 1.803 0 1 1-.99-3.467l2.31-.66a2.25 2.25 0 0 0 1.632-2.163Zm0 0V2.25L9 5.25v10.303m0 0v3.75a2.25 2.25 0 0 1-1.632 2.163l-1.32.377a1.803 1.803 0 0 1-.99-3.467l2.31-.66A2.25 2.25 0 0 0 9 15.553Z"
                    />
                </svg>
                <h3 class="mt-2 text-sm font-semibold text-slate-900 dark:text-white">
                    Aucune chanson dans votre bibliothèque
                </h3>
                <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                    Commencez par ajouter votre première chanson.
                </p>
            </div>
        {/if}
    </div>
</div>