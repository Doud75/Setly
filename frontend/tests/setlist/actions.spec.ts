import { test, expect, type Page } from '@playwright/test';

async function login(page: Page, user: string, pass: string) {
    await page.goto('/login');
    await page.getByLabel("Nom d'utilisateur").fill(user);
    await page.locator('#password').fill(pass);
    await page.getByRole('button', { name: 'Se connecter' }).click();
    await page.waitForURL('/');
}

async function createSetlist(page: Page, name: string): Promise<{ id: number; name: string }> {
    const setlistRes = await page.request.post('/api/setlist', {
        data: { name, color: '#00ffff' }
    });
    expect(setlistRes.ok()).toBeTruthy();
    const setlist = await setlistRes.json();
    return { id: setlist.id, name: setlist.name };
}

test.describe.serial('Setlist Admin Actions [As Admin]', () => {
    let setlistId: number;
    let setlistName: string;

    test.beforeEach(async ({ page }) => {
        await login(page, 'testuser', 'password123');
        const uniqueName = `Setlist for Actions Test ${Date.now()}`;
        const createdSetlist = await createSetlist(page, uniqueName);
        setlistId = createdSetlist.id;
        setlistName = createdSetlist.name;
    });

    test('should allow an admin to archive and unarchive a setlist', async ({ page }) => {
        await page.goto(`/setlist/${setlistId}`);
        await page.getByRole('button', { name: "Ouvrir le menu d'actions" }).click();
        await page.getByRole('menuitem', { name: 'Archiver' }).click();

        await expect(page.getByText('Archivée')).toBeVisible();

        await page.goto('/');
        await expect(page.getByRole('link', { name: setlistName })).toBeHidden();
        await page.getByRole('button', { name: 'Archivées' }).click();
        await expect(page.getByRole('link', { name: setlistName })).toBeVisible();

        await page.goto(`/setlist/${setlistId}`);
        await page.getByRole('button', { name: "Ouvrir le menu d'actions" }).click();
        await page.getByRole('menuitem', { name: 'Désarchiver' }).click();
        await expect(page.getByText('Archivée')).toBeHidden();

        await page.goto('/');
        await expect(page.getByRole('link', { name: setlistName })).toBeVisible();
        await page.getByRole('button', { name: 'Archivées' }).click();
        await expect(page.getByRole('link', { name: setlistName })).toBeHidden();
    });

    test('should allow an admin to delete a setlist after confirmation', async ({ page }) => {
        await page.goto(`/setlist/${setlistId}`);

        await page.getByRole('button', { name: "Ouvrir le menu d'actions" }).click();
        await page.getByRole('menuitem', { name: 'Supprimer' }).click();

        const modal = page.getByRole('dialog');
        await expect(modal).toBeVisible();
        await expect(modal).toContainText(`Êtes-vous sûr de vouloir supprimer la setlist "${setlistName}" ?`);

        await modal.getByRole('button', { name: 'Annuler' }).click();
        await expect(modal).toBeHidden();
        await expect(page.url()).toContain(`/setlist/${setlistId}`);

        await page.getByRole('button', { name: "Ouvrir le menu d'actions" }).click();
        await page.getByRole('menuitem', { name: 'Supprimer' }).click();
        await page.getByRole('button', { name: 'Confirmer la suppression' }).click();

        await page.waitForURL('/');
        await expect(page.getByRole('link', { name: setlistName })).toBeHidden();
        await page.getByRole('button', { name: 'Archivées' }).click();
        await expect(page.getByRole('link', { name: setlistName })).toBeHidden();
    });
});

test.describe('Setlist Admin Actions [As Member]', () => {
    let setlistId: number;

    test.beforeAll(async ({ browser }) => {
        const page = await browser.newPage();
        await login(page, 'testuser', 'password123');
        const createdSetlist = await createSetlist(page, `Setlist for Member View ${Date.now()}`);
        setlistId = createdSetlist.id;
        await page.request.post(`/api/setlist/${setlistId}/items`, {
            data: { item_type: 'song', item_id: 1 }
        });
        await page.close();
    });

    test('should hide delete but allow content actions for a non-admin user', async ({ page }) => {
        await login(page, 'memberuser', 'password123');
        await page.goto(`/setlist/${setlistId}`);

        await page.getByRole('button', { name: "Ouvrir le menu d'actions" }).click();
        // Un membre peut désormais dupliquer, éditer les infos et archiver...
        await expect(page.getByRole('menuitem', { name: 'Dupliquer' })).toBeVisible();
        await expect(page.getByRole('menuitem', { name: 'Modifier les infos' })).toBeVisible();
        await expect(page.getByRole('menuitem', { name: 'Archiver' })).toBeVisible();
        // ...mais pas supprimer la setlist (réservé aux admins).
        await expect(page.getByRole('menuitem', { name: 'Supprimer' })).toBeHidden();
    });

    test('should allow a member to duplicate a setlist', async ({ page }) => {
        await login(page, 'memberuser', 'password123');
        await page.goto(`/setlist/${setlistId}`);

        const newName = `Copie membre ${Date.now()}`;
        await page.getByRole('button', { name: "Ouvrir le menu d'actions" }).click();
        await page.getByRole('menuitem', { name: 'Dupliquer' }).click();
        await expect(page.getByRole('dialog')).toBeVisible();
        await page.getByLabel('Nouveau nom').fill(newName);
        await page.getByRole('button', { name: 'Créer la copie' }).click();

        await page.waitForURL(/\/setlist\/\d+/);
        await expect(page).not.toHaveURL(`/setlist/${setlistId}`);
        await expect(page.getByRole('heading', { name: newName })).toBeVisible();
    });

    test('should hide the remove-item button for a non-admin user', async ({ page }) => {
        await login(page, 'memberuser', 'password123');
        await page.goto(`/setlist/${setlistId}`);

        // L'élément est visible pour le membre...
        const items = page.locator('ul[data-testid="setlist-items"] > li');
        await expect(items.first()).toBeVisible();
        // ...mais pas le bouton pour le retirer (réservé aux admins).
        await expect(page.getByRole('button', { name: "Supprimer l'élément" })).toHaveCount(0);
    });

    test('should hide the delete-song button for a non-admin user', async ({ page }) => {
        await login(page, 'memberuser', 'password123');
        await page.goto('/song');

        const song = page.locator('li', { hasText: 'Song Title 1' });
        await expect(song).toBeVisible();
        // La suppression de chanson est réservée aux admins.
        await expect(song.getByRole('button', { name: 'Supprimer Song Title 1' })).toHaveCount(0);
    });
});