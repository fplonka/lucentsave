<script lang="ts">
	import { page } from '$app/stores';
	import type { Post } from '$lib/types';
	import { posts } from '../../../stores';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';

	let query: string = '';
	let searchResultPosts: Post[] = [...$posts];

	$: {
		if (query == '') {
			searchResultPosts = [...$posts];
		}
	}

	const getHostname = (url: string) => {
		try {
			return new URL(url).hostname;
		} catch (_) {
			return 'Invalid URL';
		}
	};

	const searchPosts = async (): Promise<void> => {
		const response = await fetch(PUBLIC_BACKEND_API_URL + `searchPosts?query=${query}`, {
			method: 'GET',
			credentials: 'include'
		});
		if (response.ok) {
			searchResultPosts = await response.json();
		} else {
			// TODO: error text?
		}
	};
</script>

<form on:submit|preventDefault={searchPosts} class="mt-5 flex items-center space-x-2">
	<input
		tabindex="1"
		type="text"
		id="url"
		bind:value={query}
		required
		class="w-full py-1 px-2 border-2 border-black"
		placeholder="Enter term to search here..."
	/>
	<input
		type="submit"
		value="Search"
		class="py-1 px-2 border-2 border-black {query !== ''
			? 'bg-black text-white hover:bg-gray-700 cursor-pointer'
			: 'bg-gray-700 text-white cursor-not-allowed'}"
		disabled={query === ''}
	/>
</form>

<div class="mt-4">
	{#each searchResultPosts as post (post.id)}
		<div class="flex justify-between items-center">
			<a href={`/post/${post.id}`} class="hover:text-gray-500">
				<h2 class="text-xl md:text-2xl font-bold">{post.title}</h2>
				<div class="text-sm block">{getHostname(post.url)}</div>
			</a>
		</div>
		{#if post.id !== searchResultPosts[searchResultPosts.length - 1].id}
			<hr class="border-black border-t-2 border-dashed my-4" />
		{/if}
	{/each}
</div>

<svelte:head>
	<title>Search - Lucentsave</title>
</svelte:head>
