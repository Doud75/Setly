<script lang="ts">
    import { toastStore } from '$lib/stores/toastStore';
    import { fly } from 'svelte/transition';

    const iconMap = {
        error: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>`,
        success: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>`,
        info: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>`,
    };

    const styleMap = {
        error: 'bg-red-600 text-white',
        success: 'bg-emerald-600 text-white',
        info: 'bg-slate-700 text-white',
    };
</script>

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none" aria-live="polite">
    {#each $toastStore as toast (toast.id)}
        <div
            class="flex items-center gap-3 rounded-lg px-4 py-3 shadow-lg text-sm font-medium max-w-sm pointer-events-auto {styleMap[toast.type]}"
            transition:fly={{ y: -12, duration: 250 }}
        >
            <!-- XSS-safe: iconMap holds only hardcoded, static SVG strings keyed by toast.type.
                 NEVER interpolate user-controlled data here. User text goes through {toast.message} below,
                 which Svelte auto-escapes. -->
            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
            {@html iconMap[toast.type]}
            <span>{toast.message}</span>
            <button
                onclick={() => toastStore.remove(toast.id)}
                class="ml-auto opacity-70 hover:opacity-100 transition-opacity"
                aria-label="Fermer"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
        </div>
    {/each}
</div>