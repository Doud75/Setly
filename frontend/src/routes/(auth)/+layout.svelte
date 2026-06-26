<script lang="ts">
    let { children } = $props();
</script>

<main class="flex min-h-svh items-center justify-center bg-slate-50 p-4 pb-[max(1rem,env(safe-area-inset-bottom))] dark:bg-slate-900">
    <div class="w-full max-w-md">
        <div class="flex justify-center mb-6">
            <svg class="h-12 w-auto text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" d="m9 9 10.5-3m0 6.553v3.75a2.25 2.25 0 0 1-1.632 2.163l-1.32.377a1.803 1.803 0 1 1-.99-3.467l2.31-.66a2.25 2.25 0 0 0 1.632-2.163Zm0 0V2.25L9 5.25v10.303m0 0v3.75a2.25 2.25 0 0 1-1.632 2.163l-1.32.377a1.803 1.803 0 0 1-.99-3.467l2.31-.66A2.25 2.25 0 0 0 9 15.553Z" />
            </svg>
        </div>
        <div class="rounded-xl bg-white p-8 shadow-2xl dark:bg-slate-800">
            {@render children?.()}
        </div>
    </div>
</main>