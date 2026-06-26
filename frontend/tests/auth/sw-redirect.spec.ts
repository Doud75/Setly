import { test, expect } from '@playwright/test';

/**
 * Régression : la PWA démarre sur `start_url: '/'`. Quand l'utilisateur est
 * déconnecté, `/` redirige côté serveur vers `/login`. Le service worker ne
 * doit PAS collapser cette redirection en réponse 200 servie sous l'URL `/` :
 * sinon la page de login s'affiche à l'URL `/`, son formulaire poste sur `/`
 * (route sans action) → `[405] POST /` et aucun message d'erreur n'apparaît.
 *
 * Ce test ouvre `/` avec le service worker actif et vérifie que la barre d'URL
 * devient bien `/login` et que la soumission affiche l'erreur attendue.
 */
test.describe('Service worker — redirection / → /login (déconnecté)', () => {
    test('ouverture sur / redirige vers /login et affiche l\'erreur de connexion', async ({ page }) => {
        // 1re visite : déclenche l'enregistrement du service worker.
        await page.goto('/');

        // Attendre que le service worker contrôle la page.
        await page.evaluate(() => navigator.serviceWorker.ready);
        await page.waitForFunction(() => navigator.serviceWorker.controller !== null);

        // Ouverture sur `/` avec le SW actif : doit atterrir sur `/login`.
        await page.goto('/');
        await expect(page).toHaveURL(/\/login$/);

        // Le formulaire doit poster sur `/login` (et non `/`) : un mauvais mot de
        // passe affiche l'erreur, l'URL reste `/login` (pas de 405 POST /).
        await page.getByLabel("Nom d'utilisateur").fill('utilisateur_inexistant');
        await page.locator('#password').fill('mauvaisMotDePasse1!');
        await page.getByRole('button', { name: 'Se connecter' }).click();

        await expect(page.getByText('Identifiant ou mot de passe incorrect.')).toBeVisible();
        await expect(page).toHaveURL(/\/login$/);
    });
});
