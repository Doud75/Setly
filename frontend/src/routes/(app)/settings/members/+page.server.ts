import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
    removeMember: async ({ request, fetch, locals }) => {
        const bandId = locals.activeBandId;
        const data = await request.formData();
        const userId = data.get('userId');

        if (!userId) {
            return fail(400, { error: 'User ID manquant.' });
        }

        const response = await fetch(`/api/bands/${bandId}/members/${userId}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const result = await response.json().catch(() => ({ error: 'Une erreur est survenue lors de la suppression.' }));
            return fail(response.status, { error: result.error });
        }

        return { removeSuccess: true, removedUserId: Number(userId) };
    },
    updateRole: async ({ request, fetch, locals }) => {
        const bandId = locals.activeBandId;
        const data = await request.formData();
        const userId = data.get('userId');
        const role = data.get('role');

        if (!userId || (role !== 'admin' && role !== 'member')) {
            return fail(400, { error: 'Requête invalide.' });
        }

        const response = await fetch(`/api/bands/${bandId}/members/${userId}/role`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ role })
        });

        if (!response.ok) {
            const result = await response.json().catch(() => ({ error: 'Une erreur est survenue lors du changement de rôle.' }));
            return fail(response.status, { error: result.error });
        }

        return { roleSuccess: true, updatedUserId: Number(userId), newRole: role };
    }
};
