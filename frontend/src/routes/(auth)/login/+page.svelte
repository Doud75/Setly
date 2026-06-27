<script lang="ts">
    import Button from '$lib/components/ui/Button.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import { enhance } from '$app/forms';
    import { navigating } from '$app/stores';

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
<<<<<<< Updated upstream
=======

    // === DEBUG TEMPORAIRE (à retirer) : collecte côté client + envoi serveur ===
    // On capte les événements bruts (capture, hors délégation Svelte) sur le bouton
    // « Se connecter » et on envoie les infos au endpoint /login/debug -> logs front.
    let btnWrap: HTMLDivElement | undefined;
    onMount(() => {
        const send = (ev: string, e?: Event) => {
            try {
                const btn = btnWrap?.querySelector('button');
                let hitCenter = 'n/a';
                const r = btn?.getBoundingClientRect();
                if (r) {
                    const h = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
                    hitCenter = h ? `${h.tagName}.${(h.getAttribute('class') || '').slice(0, 30)}` : 'null';
                }
                const tgt = e?.target as HTMLElement | undefined;
                const formEl = btnWrap?.closest('form') as HTMLFormElement | null;
                const body = JSON.stringify({
                    ev,
                    path: location.pathname,
                    formActionAttr: formEl?.getAttribute('action') ?? '(none→current url)',
                    formActionResolved: formEl?.action ?? 'n/a',
                    target: tgt ? `${tgt.tagName}.${(tgt.getAttribute('class') || '').slice(0, 30)}` : 'n/a',
                    hitCenter,
                    btnDisabled: btn?.disabled ?? null,
                    navigating: $navigating?.type ?? 'null',
                    vh: window.innerHeight,
                    vvh: window.visualViewport?.height ?? null,
                    ua: navigator.userAgent
                });
                const blob = new Blob([body], { type: 'application/json' });
                if (!navigator.sendBeacon('/login/debug', blob)) {
                    fetch('/login/debug', { method: 'POST', body, headers: { 'content-type': 'application/json' }, keepalive: true });
                }
            } catch { /* ignore */ }
        };
        send('load');
        const types = ['touchstart', 'pointerup', 'click'] as const;
        const handler = (e: Event) => send(e.type, e);
        for (const t of types) btnWrap?.addEventListener(t, handler, { capture: true, passive: true });
        const formEl = btnWrap?.closest('form');
        const onSubmit = (e: Event) => send('submit', e);
        formEl?.addEventListener('submit', onSubmit, { capture: true });
        return () => {
            for (const t of types) btnWrap?.removeEventListener(t, handler, { capture: true });
            formEl?.removeEventListener('submit', onSubmit, { capture: true });
        };
    });
>>>>>>> Stashed changes
</script>

<div class="space-y-6">
    <div>
        <h2 class="text-center text-2xl font-bold leading-9 tracking-tight text-slate-900 dark:text-white">
            Se connecter
        </h2>
    </div>

    <form method="POST" action="/login" use:enhance class="space-y-6">
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

        <Button isLoading={$navigating?.type === 'form'}>
            {#if $navigating?.type === 'form'}
                Connexion...
            {:else}
                Se connecter
            {/if}
        </Button>
    </form>

    <p class="mt-8 text-center text-sm text-slate-500 dark:text-slate-400">
        Pas de compte ? <a href="/signup" class="font-semibold leading-6 text-indigo-500 hover:text-indigo-400">Créez un groupe</a>.
    </p>
</div>