<script lang="ts">
    import {page} from '$app/stores';
    import {calculateTotalDuration, formatDuration, getSongNumber} from '$lib/utils/utils';
    import {dragHandleZone} from 'svelte-dnd-action';
    import {enhance} from '$app/forms';
    import type {ActionData, PageData} from './$types';
    import {generateSetlistPdf, generateLivePdf} from '$lib/utils/pdfGenerator';
    import SetlistItem from '$lib/components/setlist/SetlistItem.svelte';
    import EditItemForm from '$lib/components/setlist/EditItemForm.svelte';
    import type {SetlistItem as SetlistItemType} from '$lib/types';
    import DuplicateSetlistForm from '$lib/components/setlist/DuplicateSetlistForm.svelte';
    import {beforeNavigate, invalidateAll} from "$app/navigation";
    import ActionDropdown from '$lib/components/ui/ActionDropdown.svelte';
    import DeleteSetlistForm from "$lib/components/setlist/DeleteSetlistForm.svelte";
    import { modalStore } from '$lib/stores/modalStore';

    let {data, form}: { data: PageData; form: ActionData } = $props();
    const setlistId = $page.params.id;
    let items = $derived(data.setlistDetails.items);

    beforeNavigate(() => {
        modalStore.close();
    });

    $effect(() => {
        if (form?.toggledArchive) {
            data.setlistDetails.is_archived = !data.setlistDetails.is_archived;
            invalidateAll();
        }
        if (form?.deleted) {
            const index = items.findIndex((item: SetlistItemType) => item.id === form.itemId);
            if (index !== -1) {
                items.splice(index, 1);
            }
        }
        if (form?.updatedSong) {
            const index = items.findIndex((item: SetlistItemType) => item.id === form.item.id);
            if (index !== -1) {
                items[index].notes = form.item.notes;
            }
        }
        if (form?.updatedInterlude) {
            const index = items.findIndex(
                (item) => item.item_type === 'interlude' && item.interlude_id === form.interlude.id
            );
            if (index !== -1) {
                const updatedItem = items[index] as SetlistItemType;
                if (updatedItem.item_type === 'interlude') {
                    updatedItem.title = form.interlude.title;
                    updatedItem.speaker = form.interlude.speaker;
                    updatedItem.duration_seconds = form.interlude.duration_seconds;
                    items[index] = updatedItem;
                }
            }

            if (form.item) {
                const index = items.findIndex((i: SetlistItemType) => i.id === form.item.id);
                if (index !== -1) {
                    items[index].notes = form.item.notes;
                }
            }
        }
    });

    const totalDurationSeconds = $derived(calculateTotalDuration(items));

    function handleDndConsider(e: CustomEvent) {
        items = e.detail.items;
    }

    function handleDndFinalize(e: CustomEvent) {
        items = e.detail.items;
        (document.getElementById('order-form') as HTMLFormElement)?.requestSubmit();
    }

    function openEditModal(item: SetlistItemType) {
        modalStore.open(EditItemForm, { item });
    }

    function openDuplicateModal() {
        modalStore.open(DuplicateSetlistForm, {
            setlistName: data.setlistDetails.name,
            setColor: data.setlistDetails.color
        });
    }

    function openDeleteModal() {
        modalStore.open(DeleteSetlistForm, {
            setlistName: data.setlistDetails.name
        });
    }

    function downloadStandardPdf() {
        generateSetlistPdf({...data.setlistDetails, items}, totalDurationSeconds);
    }

    function downloadLivePdf() {
        generateLivePdf({...data.setlistDetails, items}, totalDurationSeconds);
    }
</script>

<div class="container mx-auto px-4 sm:px-6">
    <header class="mb-8">
        <div>
            <div class="flex flex-nowrap items-center justify-between gap-4">
                <div class="flex min-w-0 items-center gap-3">
					<span
                            class="block h-5 w-5 flex-shrink-0 rounded-full"
                            style="background-color: {data.setlistDetails.color};"
                    ></span>
                    <h1 class="truncate text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
                        {data.setlistDetails.name}
                    </h1>
                    {#if data.setlistDetails.is_archived}
						<span class="flex-shrink-0 rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-800 dark:bg-slate-700 dark:text-slate-200">
                            Archivée
                        </span>
                    {/if}
                </div>
                <ActionDropdown>
                    {#snippet children({ close })}
                        <div class="py-1" role="none">
                            <a
                                href="/setlist/{setlistId}/add"
                                class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
                                role="menuitem"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
                                    <path d="M10 4a.75.75 0 0 1 .75.75v4.5h4.5a.75.75 0 0 1 0 1.5h-4.5v4.5a.75.75 0 0 1-1.5 0v-4.5h-4.5a.75.75 0 0 1 0-1.5h4.5v-4.5A.75.75 0 0 1 10 4Z" />
                                </svg>
                                Ajouter un item
                            </a>
                            <a
                                href="/setlist/{setlistId}/edit"
                                onclick={close}
                                class="flex w-full items-center gap-2 px-4 py-2 text-sm text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
                                role="menuitem"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
                                    <path d="M15.364 2.636a2 2 0 0 1 2.828 2.828l-9.9 9.9-3.182.354a.5.5 0 0 1-.556-.556l.354-3.182 9.9-9.9Zm-2.12 2.122-8.607 8.606-.202 1.818 1.818-.202 8.607-8.606-1.616-1.616Z" />
                                </svg>
                                Modifier les infos
                            </a>
                            <button
                                onclick={() => {
                                    openDuplicateModal();
                                    close();
                                }}
                                class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
                                role="menuitem"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
                                    <path d="M7 3.5A1.5 1.5 0 0 1 8.5 2h3.879a1.5 1.5 0 0 1 1.06.44l3.122 3.121A1.5 1.5 0 0 1 17 6.621V16.5a1.5 1.5 0 0 1-1.5 1.5h-7A1.5 1.5 0 0 1 7 16.5v-13Z" />
                                    <path d="M5 5.5A1.5 1.5 0 0 1 6.5 4h1V3H6.5A2.5 2.5 0 0 0 4 5.5v11A2.5 2.5 0 0 0 6.5 19h7a2.5 2.5 0 0 0 2.5-2.5v-1h1v1A3.5 3.5 0 0 1 13.5 20h-7A3.5 3.5 0 0 1 3 16.5v-11A3.5 3.5 0 0 1 6.5 2h1V4H5V5.5Z" />
                                </svg>
                                Dupliquer
                            </button>
                            <button
                                onclick={() => {
                                    downloadStandardPdf();
                                    close();
                                }}
                                class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
                                role="menuitem"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5"><path d="M10.75 2.75a.75.75 0 0 0-1.5 0v8.614L6.295 8.235a.75.75 0 1 0-1.09 1.03l4.25 4.5a.75.75 0 0 0 1.09 0l4.25-4.5a.75.75 0 0 0-1.09-1.03l-2.955 3.129V2.75Z" /><path d="M3.5 12.75a.75.75 0 0 0-1.5 0v2.5A2.75 2.75 0 0 0 4.75 18h10.5A2.75 2.75 0 0 0 18 15.25v-2.5a.75.75 0 0 0-1.5 0v2.5c0 .69-.56 1.25-1.25 1.25H4.75c-.69 0-1.25-.56-1.25-1.25v-2.5Z" /></svg>
                                PDF (Détail)
                            </button>
                            <button
                                onclick={() => {
                                    downloadLivePdf();
                                    close();
                                }}
                                class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
                                role="menuitem"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5"><path d="M10.75 2.75a.75.75 0 0 0-1.5 0v8.614L6.295 8.235a.75.75 0 1 0-1.09 1.03l4.25 4.5a.75.75 0 0 0 1.09 0l4.25-4.5a.75.75 0 0 0-1.09-1.03l-2.955 3.129V2.75Z" /><path d="M3.5 12.75a.75.75 0 0 0-1.5 0v2.5A2.75 2.75 0 0 0 4.75 18h10.5A2.75 2.75 0 0 0 18 15.25v-2.5a.75.75 0 0 0-1.5 0v2.5c0 .69-.56 1.25-1.25 1.25H4.75c-.69 0-1.25-.56-1.25-1.25v-2.5Z" /></svg>
                                PDF (Live)
                            </button>
                                <div class="border-t border-slate-200 py-1 dark:border-slate-700">
                                    <form method="POST" action="?/toggleArchiveStatus" use:enhance>
                                        <input type="hidden" name="is_archived" value={data.setlistDetails.is_archived} />
                                        <button
                                            type="submit"
                                            class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-100 disabled:opacity-50 dark:text-slate-200 dark:hover:bg-slate-700"
                                            role="menuitem"
                                        >
                                            {#if data.setlistDetails.is_archived}
                                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5"><path d="M3.5 3.75A1.75 1.75 0 0 1 5.25 2h9.5A1.75 1.75 0 0 1 16.5 3.75v12.5A1.75 1.75 0 0 1 14.75 18h-9.5A1.75 1.75 0 0 1 3.5 16.25V3.75ZM8.56 8.28a.75.75 0 0 0-1.06 1.06L9.44 11.5l-1.94 2.16a.75.75 0 1 0 1.12 1.004l2.5-2.75a.75.75 0 0 0 0-1.004l-2.5-2.75a.75.75 0 0 0-1.06 0Z" /></svg>
                                                Désarchiver
                                            {:else}
                                                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
                                                    <path d="M3.5 3.75A1.75 1.75 0 0 1 5.25 2h9.5A1.75 1.75 0 0 1 16.5 3.75v12.5A1.75 1.75 0 0 1 14.75 18h-9.5A1.75 1.75 0 0 1 3.5 16.25V3.75ZM8.25 7a.75.75 0 0 0 0 1.5h3.5a.75.75 0 0 0 0-1.5h-3.5Z" />
                                                </svg>
                                                Archiver
                                            {/if}
                                        </button>
                                    </form>
                                </div>
                                {#if data.user?.role === 'admin'}
                                <div class="border-t border-slate-200 py-1 dark:border-slate-700">
                                    <button
                                        onclick={() => {
                                            openDeleteModal();
                                            close();
                                        }}
                                        class="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-red-700 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10"
                                        role="menuitem"
                                    >
                                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="h-5 w-5">
                                            <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.134-2.033-2.134H8.033C6.91 2.75 6 3.704 6 4.874v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
                                        </svg>
                                        Supprimer
                                    </button>
                                </div>
                            {/if}
                        </div>
                    {/snippet}
                </ActionDropdown>

            </div>
            <div class="mt-2 flex items-center gap-4 text-sm text-slate-500 dark:text-slate-400">
                <a href="/" class="hover:underline">&larr; Retour</a>
                <span>&bull;</span>
                <span
                >Durée totale : <span class="font-semibold"
                >{formatDuration(totalDurationSeconds)}</span
                ></span
                >
            </div>
        </div>
    </header>

    <div class="rounded-xl bg-white p-6 shadow-lg dark:bg-slate-800">
        {#if items && items.length > 0}
            <form id="order-form" method="POST" action="?/updateOrder" use:enhance>
                <input type="hidden" name="itemIds" value={JSON.stringify(items.map((item) => item.id))}/>
            </form>

            <ul
                    data-testid="setlist-items"
                    class="divide-y divide-slate-200 dark:divide-slate-700"
                    use:dragHandleZone={{ items: items, flipDurationMs: 300 }}
                    onconsider={handleDndConsider}
                    onfinalize={handleDndFinalize}
            >
                {#each items as item (item.id)}
                    <SetlistItem
                            {item}
                            songNumber={getSongNumber(item, items)}
                            onEdit={openEditModal}
                    />
                {/each}
            </ul>
        {:else}
            <div class="py-12 text-center">
                <svg
                        class="mx-auto h-12 w-12 text-slate-400"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke-width="1.5"
                        stroke="currentColor"
                >
                    <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="m9 9 10.5-3m0 6.553v3.75a2.25 2.25 0 0 1-1.632 2.163l-1.32.377a1.803 1.803 0 0 1-.99-3.467l2.31-.66a2.25 2.25 0 0 0 1.632-2.163Zm0 0V2.25L9 5.25v10.303m0 0v3.75a2.25 2.25 0 0 1-1.632 2.163l-1.32.377a1.803 1.803 0 0 1-.99-3.467l2.31-.66A2.25 2.25 0 0 0 9 15.553Z"
                    />
                </svg>
                <h3 class="mt-2 text-sm font-semibold text-slate-900 dark:text-white">
                    Cette setlist est vide
                </h3>
                <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                    Commencez par ajouter votre premier élément.
                </p>
            </div>
        {/if}
    </div>
</div>