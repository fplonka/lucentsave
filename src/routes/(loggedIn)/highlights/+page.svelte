<script lang="ts">
	import PostBody from '../PostBody.svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	let highlightsWithTitles = data.highlights;
</script>

<div class="mt-4">
	{#each highlightsWithTitles as highlight (highlight.id)}
		<div>
			<a
				href={`/post/${highlight.postId}`}
				class="text-xl md:text-2xl font-bold block hover:text-gray-500">{highlight.title}</a
			>
			<a href={`/post/${highlight.postId}#${highlight.id}`}>
				<PostBody classes="hover:!text-gray-500">
					{@html highlight.text}
				</PostBody>
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
