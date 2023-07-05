<script lang="ts">
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import type { Post } from '../[listType]/+page.server';

	export let data: PageData;

	let searchQuery: string = '';
	$: searchResultPosts = data.posts.filter(
		(p: Post) =>
			p.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
			p.body.toLowerCase().includes(searchQuery.toLowerCase()) ||
			p.url.toLowerCase().includes(searchQuery.toLowerCase())
	);

	const getHostname = (url: string) => {
		try {
			return new URL(url).hostname;
		} catch (_) {
			return 'Invalid URL';
		}
	};
</script>

<div class="mt-5 flex items-center space-x-2">
	<input
		tabindex="1"
		type="text"
		id="query"
		bind:value={searchQuery}
		required
		class="w-full py-1 px-2 border-2 border-black"
		placeholder="Enter search term here..."
	/>
</div>

<div class="mt-4">
	{#each searchResultPosts as post (post.id)}
		<div class="flex justify-between items-center">
			<a href={`/post/${post.id}`} class="hover:text-gray-500">
				<div class="text-2xl font-bold block">{post.title}</div>
				<div class="text-sm block">{getHostname(post.url)}</div>
			</a>
		</div>
		{#if post.id !== data.posts[data.posts.length - 1].id}
			<hr class="border-black border-t-2 border-dashed my-4" />
		{/if}
	{/each}
</div>

{#if data.posts.length == 0}
	<div class="mt-4 italic">Nothing {$page.url.pathname.substring(1)} yet...</div>
{/if}
