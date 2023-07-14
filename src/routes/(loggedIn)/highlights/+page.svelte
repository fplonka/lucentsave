<script lang="ts">
	import { onMount } from 'svelte';
	import PostBody from '../PostBody.svelte';
	import type { PageData } from './$types';
	import { goto } from '$app/navigation';

	export let data: PageData;

	let highlightsWithTitles = data.highlights;

	let highlightIdToPostId: { [key: string]: number } = {};
	highlightsWithTitles.forEach((highlight) => {
		highlightIdToPostId[highlight.id] = highlight.postId;
	});

	function reloadEventListeners() {
		// Select all highlighted spans
		let highlights = Array.from(document.querySelectorAll(`span[data-highlight-id]`));

		// Remove any previous listeners (to avoid duplication if this function is called more than once)
		highlights.forEach((highlight) => {
			highlight.removeEventListener('mouseenter', handleMouseEnter);
			highlight.removeEventListener('mouseleave', handleMouseLeave);

			highlight.addEventListener('click', (event) => {
				event.preventDefault(); // prevent any default action
				let highlightID = (event.currentTarget as HTMLElement).dataset.highlightId;
				let postId = highlightIdToPostId[highlightID!];
				goto(`/post/${postId}#${highlightID}`);
			});
		});

		// Attach new event listeners
		highlights.forEach((highlight) => {
			highlight.addEventListener('mouseenter', handleMouseEnter);
			highlight.addEventListener('mouseleave', handleMouseLeave);

			highlight.addEventListener('click', (event) => {
				event.preventDefault(); // prevent any default action
				let highlightID = (event.currentTarget as HTMLElement).dataset.highlightId;
				let postId = highlightIdToPostId[highlightID!];
				goto(`/post/${postId}#${highlightID}`);
			});
		});
	}

	// Event handler for mouseenter
	function handleMouseEnter(this: HTMLElement) {
		let highlightID = this.dataset.highlightId;
		let highlights = document.querySelectorAll(`span[data-highlight-id="${highlightID}"]`);
		highlights.forEach((highlight) => highlight.classList.add('bg-yellow-300'));
	}

	// Event handler for mouseleave
	function handleMouseLeave(this: HTMLElement) {
		let highlightID = this.dataset.highlightId;
		let highlights = document.querySelectorAll(`span[data-highlight-id="${highlightID}"]`);
		highlights.forEach((highlight) => highlight.classList.remove('bg-yellow-300'));
	}

	onMount(() => {
		reloadEventListeners();
	});
</script>

<div class="mt-4">
	{#each highlightsWithTitles as highlight (highlight.id)}
		<div>
			<a
				href={`/post/${highlight.postId}`}
				class="text-xl md:text-2xl font-bold block hover:text-gray-500">{highlight.title}</a
			>
			<PostBody classes="">
				{@html highlight.text}
			</PostBody>
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
