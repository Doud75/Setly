// frontend/tests/settings/members.spec.ts

import { test, expect, type Page } from '@playwright/test';

async function login(page: Page) {
    await page.goto('/login');
    await page.getByLabel("Nom d'utilisateur").fill('testuser');
    await page.locator('#password').fill('password123');
    await page.getByRole('button', { name: 'Se connecter' }).click();
    await page.waitForURL('/');
}

test.describe.serial('Settings - Members Page (Admin)', () => {
    test.beforeEach(async ({ page }) => {
        await login(page);
        await page.goto('/settings/members');
    });

    test('should display the list of current members', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Gérer les membres du groupe' })).toBeVisible();

        const adminUserRow = page.locator('li', { hasText: 'testuser' });
        await expect(adminUserRow).toBeVisible();
        await expect(adminUserRow.getByText('admin')).toBeVisible();

        const memberUserRow = page.locator('li', { hasText: 'memberuser' });
        await expect(memberUserRow).toBeVisible();
        await expect(memberUserRow.getByText('member', { exact: true })).toBeVisible();
    });

    test('should promote a member to admin and demote back', async ({ page }) => {
        const memberUserRow = page.locator('li', { hasText: 'memberuser' });
        await expect(memberUserRow.getByText('member', { exact: true })).toBeVisible();

        // Promotion member -> admin
        await memberUserRow.getByRole('button', { name: 'Promouvoir admin' }).click();
        await expect(memberUserRow.getByText('admin', { exact: true })).toBeVisible();
        await page.reload();
        await expect(page.locator('li', { hasText: 'memberuser' }).getByText('admin', { exact: true })).toBeVisible();

        // Rétrogradation admin -> member (restaure l'état pour les tests suivants)
        const promotedRow = page.locator('li', { hasText: 'memberuser' });
        await promotedRow.getByRole('button', { name: 'Rétrograder' }).click();
        await expect(promotedRow.getByText('member', { exact: true })).toBeVisible();
        await page.reload();
        await expect(page.locator('li', { hasText: 'memberuser' }).getByText('member', { exact: true })).toBeVisible();
    });

    test('should not allow demoting the last admin', async ({ page }) => {
        const adminUserRow = page.locator('li', { hasText: 'testuser' });
        // testuser est le seul admin -> le bouton Rétrograder doit être désactivé.
        await expect(adminUserRow.getByRole('button', { name: 'Rétrograder' })).toBeDisabled();
    });

    test('should successfully remove a member', async ({ page }) => {
        // Crée un membre jetable pour ne PAS muter l'état partagé (memberuser),
        // qui doit rester dans le groupe pour les autres fichiers de tests.
        // Le backend rattache le membre au groupe actif (X-Band-ID), l'id d'URL est ignoré.
        const disposable = `dispo_${Date.now()}`;
        const inviteRes = await page.request.post('/api/bands/1/members', {
            data: { username: disposable, password: 'Password123!' }
        });
        expect(inviteRes.ok()).toBeTruthy();

        await page.reload();
        const disposableRow = page.locator('li', { hasText: disposable });
        await expect(disposableRow).toBeVisible();

        await disposableRow.getByRole('button', { name: `Supprimer ${disposable}` }).click();
        await expect(disposableRow).toBeHidden();

        await page.reload();
        await expect(page.locator('li', { hasText: disposable })).toBeHidden();
        // memberuser est intact, testuser toujours là.
        await expect(page.locator('li', { hasText: 'memberuser' })).toBeVisible();
        await expect(page.locator('li', { hasText: 'testuser' })).toBeVisible();
    });

    test('should display the invitation link section', async ({ page }) => {
        await expect(page.getByRole('heading', { name: 'Inviter par lien' })).toBeVisible();
        await expect(page.getByRole('button', { name: 'Générer un lien' })).toBeVisible();
    });
});