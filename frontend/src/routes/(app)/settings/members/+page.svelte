<script lang="ts">
    import { enhance } from '$app/forms';
    import type { ActionData, PageData } from './$types';

    let inviteLink = $state('');
    let inviteExpiry = $state('');
    let inviteLinkError = $state('');
    let isGenerating = $state(false);
    let copied = $state(false);

    async function generateInviteLink() {
        isGenerating = true;
        inviteLink = '';
        inviteLinkError = '';
        inviteExpiry = '';
        try {
            const bandId = data.activeBandId;
            const res = await fetch(`/api/bands/${bandId}/invitations`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ role: 'member' })
            });
            if (!res.ok) {
                inviteLinkError = "Impossible de générer le lien d'invitation.";
                return;
            }
            const result = await res.json();
            inviteLink = `${window.location.origin}/join/${result.token}`;
            if (result.expires_at) {
                inviteExpiry = new Date(result.expires_at).toLocaleDateString('fr-FR', {
                    day: 'numeric', month: 'long', year: 'numeric'
                });
            }
        } catch {
            inviteLinkError = "Impossible de générer le lien d'invitation.";
        } finally {
            isGenerating = false;
        }
    }

    async function copyLink() {
        if (!inviteLink) return;
        await navigator.clipboard.writeText(inviteLink);
        copied = true;
        window.setTimeout(() => { copied = false; }, 2000);
    }


    let { data, form }: { data: PageData; form: ActionData } = $props();

    let members = $derived(data.members);

    $effect(() => {
        if (form?.removeSuccess) {
            const index = members.findIndex((m) => m.id === form.removedUserId);
            if (index !== -1) {
                members.splice(index, 1);
            }
        }
        if (form?.roleSuccess && form.newRole) {
            const index = members.findIndex((m) => m.id === form.updatedUserId);
            if (index !== -1) {
                members[index] = { ...members[index], role: form.newRole };
                members = members;
            }
        }
    });
</script>

<div class="container mx-auto px-4 sm:px-6">
    <header class="mb-8">
        <h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">
            Gérer les membres du groupe
        </h1>
        <div class="mt-2 flex items-center gap-4 text-sm text-slate-500 dark:text-slate-400">
            <a href="/" class="hover:underline">&larr; Retour à l'accueil</a>
        </div>
    </header>

    <div class="grid grid-cols-1 gap-12 lg:grid-cols-3">
        <div class="lg:col-span-2">
            <div class="rounded-xl bg-white p-6 shadow-lg dark:bg-slate-800">
                <h2 class="text-xl font-semibold text-slate-800 dark:text-slate-100">
                    Membres ({members.length})
                </h2>
                {#if members.length > 0}
                    <ul class="mt-4 divide-y divide-slate-200 dark:divide-slate-700">
                        {#each members as member (member.id)}
                            <li class="flex items-center justify-between gap-3 py-4">
                                <div>
                                    <p class="font-semibold text-slate-800 dark:text-slate-100">
                                        {member.username}
                                    </p>
                                    <span
                                            class="mt-1 inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset {member.role ===
										'admin'
											? 'bg-blue-50 text-blue-700 ring-blue-600/20 dark:bg-blue-500/10 dark:text-blue-400 dark:ring-blue-500/20'
											: 'bg-slate-50 text-slate-600 ring-slate-500/20 dark:bg-slate-500/10 dark:text-slate-400 dark:ring-slate-500/20'}"
                                    >
										{member.role}
									</span>
                                </div>

                                <div class="flex items-center gap-1">
                                <form method="POST" action="?/updateRole" use:enhance>
                                    <input type="hidden" name="userId" value={member.id} />
                                    <input type="hidden" name="role" value={member.role === 'admin' ? 'member' : 'admin'} />
                                    <button
                                            type="submit"
                                            class="rounded-md px-2 py-1 text-xs font-medium text-slate-600 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50 dark:text-slate-300 dark:hover:bg-slate-700"
                                            disabled={member.role === 'admin' &&
											members.filter((m) => m.role === 'admin').length <= 1}
                                    >
                                        {member.role === 'admin' ? 'Rétrograder' : 'Promouvoir admin'}
                                    </button>
                                </form>
                                <form method="POST" action="?/removeMember" use:enhance>
                                    <input type="hidden" name="userId" value={member.id} />
                                    <button
                                            type="submit"
                                            class="rounded-md p-2 text-slate-400 hover:bg-red-50 hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-50 dark:text-slate-400 dark:hover:bg-red-500/10 dark:hover:text-red-400"
                                            aria-label="Supprimer {member.username}"
                                            disabled={member.role === 'admin' &&
											members.filter((m) => m.role === 'admin').length <= 1}
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
                                </div>
                            </li>
                        {/each}
                    </ul>
                {:else}
                    <p class="mt-4 text-sm text-slate-500 dark:text-slate-400">Aucun membre dans ce groupe.</p>
                {/if}
            </div>
        </div>

        <div class="lg:col-span-1 space-y-6">
            <div class="rounded-xl bg-white p-6 shadow-lg dark:bg-slate-800">
                <h2 class="text-xl font-semibold text-slate-800 dark:text-slate-100">Inviter par lien</h2>
                <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                    Générez un lien valable 3 jours à envoyer à vos musiciens.
                </p>
                <div class="mt-4 space-y-3">
                    <button
                        type="button"
                        onclick={generateInviteLink}
                        disabled={isGenerating}
                        class="flex w-full items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        {#if isGenerating}
                            <svg class="h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                            </svg>
                            Génération...
                        {:else}
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
                            </svg>
                            Générer un lien
                        {/if}
                    </button>

                    {#if inviteLinkError}
                        <p class="text-sm text-red-500">{inviteLinkError}</p>
                    {/if}

                    {#if inviteLink}
                        <div class="space-y-2">
                            <div class="flex gap-2">
                                <input
                                    type="text"
                                    readonly
                                    value={inviteLink}
                                    class="min-w-0 flex-1 rounded-lg border border-slate-300 bg-slate-50 px-3 py-2 text-xs text-slate-700 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-slate-300"
                                />
                                <button
                                    type="button"
                                    onclick={copyLink}
                                    class="shrink-0 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200 dark:hover:bg-slate-600"
                                    aria-label="Copier le lien"
                                >
                                    {#if copied}
                                        ✓
                                    {:else}
                                        Copier
                                    {/if}
                                </button>
                            </div>
                            {#if inviteExpiry}
                                <p class="text-xs text-slate-400 dark:text-slate-500">
                                    Valide jusqu'au {inviteExpiry}
                                </p>
                            {/if}
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    </div>
</div>