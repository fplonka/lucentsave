<script lang="ts">
	import { posts, postsLoaded } from '../../stores';
	import type { PageData } from './$types';
	import type { Post } from '../[listType]/+page.server';
	import { get } from 'svelte/store';

	export let data: PageData;

	let postList: Post[] = get(posts);

	console.log('posts: ');
	postList.forEach(console.log);

	// Add the post title to each highlight
	let highlightsWithTitles = data.highlights.map((highlight) => {
		console.log(`searching for ${highlight.postId}`);
		let post = postList.find((post) => post.id === highlight.postId);
		return {
			...highlight,
			title: post ? post.title : 'Post not found'
		};
	});
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

<svelte:head>
	<title>Highlights - Lucentsave</title>
</svelte:head>
