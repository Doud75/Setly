<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import PWABadge from '$lib/components/PWABadge.svelte';
	import OfflineIndicator from '$lib/components/OfflineIndicator.svelte';
	import { modalStore } from '$lib/stores/modalStore';
	import Modal from '$lib/components/ui/Modal.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<PWABadge/>
<OfflineIndicator />

{@render children?.()}

{#if $modalStore.isOpen && $modalStore.component}
	<Modal isOpen={$modalStore.isOpen} onClose={modalStore.close}>
		{@const Component = $modalStore.component}
		<Component {...$modalStore.props} close={modalStore.close} />
	</Modal>
{/if}