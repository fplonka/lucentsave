<script lang="ts">
	import type { PageData } from './$types';

	export let data: PageData;

	let highlightsWithTitles = data.highlights;
</script>

<div class="mt-4">
	{#each highlightsWithTitles as highlight (highlight.id)}
		<div class="flex justify-between items-center">
			<a href={`/post/${highlight.postId}#${highlight.id}`} class="hover:text-gray-500">
				<div class="text-xl md:text-2xl font-bold block">{highlight.title}</div>
				<div class="text-sm block">{highlight.text}</div>
			</a>
		</div>
		{#if highlight.id !== data.highlights[data.highlights.length - 1].id}
			<hr class="border-black border-t-2 border-dashed my-4" />
		{/if}
	{/each}
</div>

{#if highlightsWithTitles.length == 0}
	<div class="mt-4 italic">Nothing highlighted yet...</div>
{/if}

<svelte:head>
	<title>Highlights - Lucentsave</title>
</svelte:head>
