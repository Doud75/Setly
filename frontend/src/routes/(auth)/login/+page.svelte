<script lang="ts">
    import Button from '$lib/components/ui/Button.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import { enhance } from '$app/forms';
    import { navigating } from '$app/stores';
    import { onMount } from 'svelte';

    type ActionData = {
        error?: string;
        code?: string;
    } | null;

    let { form, data }: { form: ActionData; data: { redirectTo?: string } } = $props();

    const errorMessages: Record<string, string> = {
        INVALID_CREDENTIALS: 'Identifiant ou mot de passe incorrect.',
        INVALID_REQUEST: 'Requête invalide. Vérifiez vos informations.',
        INTERNAL_ERROR: 'Une erreur serveur s\'est produite. Réessayez dans un instant.',
    };

    const errorMessage = $derived(
        form
            ? (form.code && errorMessages[form.code]) ?? form.error ?? 'Une erreur inattendue s\'est produite.'
            : null
    );

    // === DEBUG TEMPORAIRE (à retirer) : collecte côté client + envoi serveur ===
    let btnWrap: HTMLDivElement | undefined;

    function send(ev: string, extra?: Record<string, unknown>) {
        try {
            const body = JSON.stringify({ ev, path: location.pathname, ...extra });
            const blob = new Blob([body], { type: 'application/json' });
            if (!navigator.sendBeacon('/login/debug', blob)) {
                fetch('/login/debug', { method: 'POST', body, headers: { 'content-type': 'application/json' }, keepalive: true });
            }
        } catch { /* ignore */ }
    }

    // Instrumente le cycle de vie de use:enhance pour voir ce qui se passe APRÈS le
    // submit : le fetch part-il ? quel résultat le serveur renvoie-t-il ? erreur ?
    type EnhanceResult = { type?: string; status?: number; location?: string };
    function enhanceLog() {
        send('enhance:submit');
        return async ({ result, update }: { result: EnhanceResult; update: () => Promise<void> }) => {
            send('enhance:result', { resultType: result?.type, status: result?.status, location: result?.location });
            await update();
        };
    }

    onMount(() => {
        send('load', { ua: navigator.userAgent });
        const types = ['touchstart', 'pointerup', 'click'] as const;
        const handler = (e: Event) => send(e.type, { target: (e.target as HTMLElement)?.tagName });
        for (const t of types) btnWrap?.addEventListener(t, handler, { capture: true, passive: true });
        const formEl = btnWrap?.closest('form');
        const onSubmit = () => send('submit');
        formEl?.addEventListener('submit', onSubmit, { capture: true });
        const onErr = (e: ErrorEvent) => send('window-error', { msg: String(e.message), src: e.filename, line: e.lineno });
        const onRej = (e: PromiseRejectionEvent) => send('unhandledrejection', { reason: String(e.reason) });
        window.addEventListener('error', onErr);
        window.addEventListener('unhandledrejection', onRej);
        return () => {
            for (const t of types) btnWrap?.removeEventListener(t, handler, { capture: true });
            formEl?.removeEventListener('submit', onSubmit, { capture: true });
            window.removeEventListener('error', onErr);
            window.removeEventListener('unhandledrejection', onRej);
        };
    });
</script>

<div class="space-y-6">
    <div>
        <h2 class="text-center text-2xl font-bold leading-9 tracking-tight text-slate-900 dark:text-white">
            Se connecter
        </h2>
    </div>

    <form method="POST" action="/login" use:enhance={enhanceLog} class="space-y-6">
        {#if data.redirectTo}
            <input type="hidden" name="redirectTo" value={data.redirectTo} />
        {/if}
        <Input label="Nom d'utilisateur" id="username" name="username" required />
        <Input label="Mot de passe" id="password" name="password" type="password" required togglePasswordVisibility={true}/>

        {#if errorMessage}
            <p class="rounded-md bg-red-50 p-3 text-center text-sm font-medium text-red-600 dark:bg-red-900/20 dark:text-red-400">
                {errorMessage}
            </p>
        {/if}

        <div bind:this={btnWrap}>
            <Button isLoading={$navigating?.type === 'form'}>
                {#if $navigating?.type === 'form'}
                    Connexion...
                {:else}
                    Se connecter
                {/if}
            </Button>
        </div>
    </form>

    <p class="mt-8 text-center text-sm text-slate-500 dark:text-slate-400">
        Pas de compte ? <a href="/signup" class="font-semibold leading-6 text-indigo-500 hover:text-indigo-400">Créez un groupe</a>.
    </p>
</div>
