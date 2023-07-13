<script>
	import Banner from '../Banner.svelte';
	import PageBody from '../PageBody.svelte';
	import '../../app.css';
</script>

<PageBody>
	<Banner />
	<slot />
</PageBody>
