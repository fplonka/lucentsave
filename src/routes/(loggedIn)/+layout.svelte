<script>
	import Navbar from '../Navbar.svelte';
	import PageBody from '../PageBody.svelte';
</script>

<Navbar>
	<PageBody>
		<slot />
	</PageBody>
</Navbar>
