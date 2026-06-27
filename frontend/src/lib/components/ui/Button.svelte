<script lang="ts">
    import type {Snippet} from "svelte";

    let {
        type = 'submit',
        isLoading = false,
        disabled = false,
        variant = 'primary',
        autoWidth = false,
        onclick,
        children
    } = $props<{
        type?: 'submit' | 'button' | 'reset';
        isLoading?: boolean;
        disabled?: boolean;
        variant?: 'primary' | 'secondary';
        autoWidth?: boolean;
        onclick?: (event: MouseEvent) => void;
        children: Snippet;
    }>();

    const baseClasses = "flex justify-center rounded-md px-4 py-2.5 text-sm font-semibold shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-slate-800 disabled:cursor-not-allowed disabled:opacity-60";

    const variants = {
        primary: 'bg-indigo-600 text-white hover:bg-indigo-500 focus:ring-indigo-500',
        secondary: 'bg-teal-600 text-white hover:bg-teal-500 focus:ring-teal-500'
    };
</script>

<button {type} disabled={isLoading || disabled} {onclick} class="{baseClasses} {variants[variant]} {autoWidth ? 'w-auto' : 'w-full'}">
    {#if isLoading}
        <svg class="h-5 w-5 animate-spin text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
    {:else}
        {@render children()}
    {/if}
</button>