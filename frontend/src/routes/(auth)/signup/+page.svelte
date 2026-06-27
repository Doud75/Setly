<script lang="ts">
    import Button from '$lib/components/ui/Button.svelte';
    import Input from '$lib/components/ui/Input.svelte';
    import { enhance } from '$app/forms';
    import { navigating } from '$app/stores';

    type ActionData = {
        error?: string;
        errors?: Record<string, string>;
        data?: {
            username?: string;
        };
    } | null;

    let { form, data }: { form: ActionData; data: { redirectTo?: string } } = $props();

    let username = $state('');
    let password = $state('');

    $effect(() => {
        if (form?.data?.username) username = form.data.username;
    });

    let hasMinLength = $derived(password.length >= 8);
    let hasUppercase = $derived(/[A-Z]/.test(password));
    let hasNumber = $derived(/[0-9]/.test(password));
    // eslint-disable-next-line no-useless-escape
    let hasSpecial = $derived(/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password));

    let usernameLengthValid = $derived(username.length >= 3 && username.length <= 50);
    let usernamePatternValid = $derived(/^[a-zA-Z0-9_]+$/.test(username));
</script>

<div class="space-y-6">
    <div>
        <h2 class="text-center text-2xl font-bold leading-9 tracking-tight text-slate-900 dark:text-white">
            Créer un compte
        </h2>
        <p class="mt-2 text-center text-sm text-slate-600 dark:text-slate-400">
            Choisissez un nom d'utilisateur et un mot de passe pour commencer.
        </p>
    </div>

    <form method="POST" use:enhance class="space-y-6">
        {#if data.redirectTo}
            <input type="hidden" name="redirectTo" value={data.redirectTo} />
        {/if}

        <div class="space-y-1">
            <Input
                    label="Nom d'utilisateur"
                    id="username"
                    name="username"
                    placeholder="votre_pseudo"
                    required
                    bind:value={username}
            />
            {#if form?.errors?.username}
                <p class="text-sm text-red-500 font-medium">{form.errors.username}</p>
            {/if}
            <ul class="text-xs text-slate-500 list-disc ml-4 space-y-1 mt-2">
                <li class={usernameLengthValid ? 'text-teal-600 dark:text-teal-400' : ''}>3-50 caractères</li>
                <li class={usernamePatternValid ? 'text-teal-600 dark:text-teal-400' : ''}>Alphanumérique & underscore uniquement</li>
            </ul>
        </div>

        <div class="space-y-1">
            <Input
                    label="Mot de passe"
                    id="password"
                    name="password"
                    type="password"
                    required
                    togglePasswordVisibility={true}
                    bind:value={password}
            />
            {#if form?.errors?.password}
                <p class="text-sm text-red-500 font-medium">{form.errors.password}</p>
            {/if}

            <ul class="text-xs text-slate-500 list-disc ml-4 space-y-1 mt-2">
                <li class={hasMinLength ? 'text-teal-600 dark:text-teal-400' : ''}>Minimum 8 caractères</li>
                <li class={hasUppercase ? 'text-teal-600 dark:text-teal-400' : ''}>Au moins 1 majuscule</li>
                <li class={hasNumber ? 'text-teal-600 dark:text-teal-400' : ''}>Au moins 1 chiffre</li>
                <li class={hasSpecial ? 'text-teal-600 dark:text-teal-400' : ''}>Au moins 1 caractère spécial</li>
            </ul>
        </div>

        {#if form?.error}
            <p class="text-center text-sm text-red-500 font-bold bg-red-50 dark:bg-red-900/10 p-2 rounded">{form.error}</p>
        {/if}

        <Button isLoading={$navigating?.type === 'form'}>
            {#if $navigating?.type === 'form'}
                Création...
            {:else}
                Créer mon compte
            {/if}
        </Button>
    </form>

    <p class="mt-8 text-center text-sm text-slate-500 dark:text-slate-400">
        Déjà un compte ? <a href="/login" class="font-semibold leading-6 text-indigo-500 hover:text-indigo-400">Se connecter</a>.
    </p>
</div>