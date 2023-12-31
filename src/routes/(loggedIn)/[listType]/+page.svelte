<script lang="ts">
	import { page } from '$app/stores';
	import { like } from '$lib/postActions';
	import type { Post } from '$lib/types';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';
	import { isSignedIn, posts } from '../../../stores';
	import { postsLoaded } from '../../../stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { capitalizeFirstLetter } from '$lib/capitalize';

	if (browser && !$postsLoaded) {
		const fetchPosts = async () => {
			const response = await fetch(PUBLIC_BACKEND_API_URL + 'getAllUserPosts', {
				credentials: 'include'
			});
			if (response.ok) {
				posts.set(await response.json());
				postsLoaded.set(true);
				isSignedIn.set(true);
			} else if (response.status == 401) {
				// Sign the user out. This is important in the weird edge case where the user has a token they think
				// is valid but something changed on the backend such that it no longer is.
				await fetch(PUBLIC_BACKEND_API_URL + 'signout', {
					method: 'POST',
					credentials: 'include'
				});
				goto('/signin');
				isSignedIn.set(false);
				postsLoaded.set(false);
			}
		};
		fetchPosts();
	}

	$: filteredPosts = filterPosts($posts, $page.url.pathname.substring(1));

	let url: string = '';
	let title: string = '';
	let body: string = '';

	$: isUrlValid = /^http(s)?:\/\/[^\s$.?#].[^\s]*$/.test(url);

	const waitingText = 'Saving, please wait...';
	let isSaving = false;
	let savingText = waitingText;

	async function savePostFromURL(event: Event) {
		isSaving = true;
		let urlToSave = url;
		url = '';
		event.preventDefault();

		const response = await fetch(PUBLIC_BACKEND_API_URL + `createPostFromURL?url=${urlToSave}`, {
			credentials: 'include'
		});
		if (!response.ok) {
			savingText = await response.text();
		} else {
			posts.set(await response.json());
			isSaving = false;
		}
	}

	const updatePost = async (post: Post) => {
		const postIndex = $posts.findIndex((p) => p.id === post.id);

		if (postIndex !== -1) {
			$posts[postIndex] = post;
		}
	};

	const likePostAndUpdate = async (post: Post) => {
		const postCopy = { ...post };

		const postIndex = $posts.findIndex((p) => p.id === post.id);
		if (postIndex !== -1) {
			$posts[postIndex].isLiked = !post.isLiked;
		}

		const udpatedPost = await like(postCopy);
		updatePost(udpatedPost);
	};

	const getHostname = (url: string) => {
		try {
			return new URL(url).hostname;
		} catch (_) {
			return 'Invalid URL';
		}
	};

	const filterPosts = (posts: Post[], path: string) => {
		switch (path) {
			case 'saved':
				return posts.filter((post) => !post.isRead);
			case 'liked':
				return posts.filter((post) => post.isLiked);
			case 'read':
				return posts.filter((post) => post.isRead);
			default:
				return posts;
		}
	};
</script>

{#if $page.url.pathname.startsWith('/saved')}
	<form on:submit={savePostFromURL} class="mt-5 flex items-center space-x-2">
		<input
			tabindex="1"
			type="text"
			id="url"
			bind:value={url}
			required
			class="w-full py-1 px-2 border-2 border-black"
			placeholder="Enter link to save here..."
		/>
		<input
			type="submit"
			value="Save"
			class="py-1 px-2 border-2 border-black {isUrlValid
				? 'bg-black text-white hover:bg-gray-700 cursor-pointer'
				: 'bg-gray-700 text-white cursor-not-allowed'}"
			disabled={!isUrlValid}
		/>
	</form>
{/if}

{#if $page.url.pathname.startsWith('/saved') && isSaving}
	<div class="flex justify-between items-center mt-4 border-black border-b-2 border-dashed pb-4">
		<div>
			<div class="text-sm block italic">{savingText}</div>
		</div>
		<div />
	</div>
{/if}

<div class="mt-4">
	{#each filteredPosts as post (post.id)}
		<div class="flex justify-between items-center">
			<a href={`/post/${post.id}`} class="hover:text-gray-500">
				<div class="text-xl md:text-2xl font-bold block">{post.title}</div>
				<div class="text-sm block">{getHostname(post.url)}</div>
			</a>
			<div>
				{#if !$page.url.pathname.startsWith('/saved')}
					<span
						role="button"
						tabindex="2"
						on:click={() => likePostAndUpdate(post)}
						on:keydown={(e) => e.key === 'Enter' && likePostAndUpdate(post)}
						class="text-black px-2 py-1 text-xl cursor-pointer hover:text-gray-500"
						>{post.isLiked ? '★' : '☆'}</span
					>
				{/if}
			</div>
		</div>
		{#if post.id !== filteredPosts[filteredPosts.length - 1].id}
			<hr class="border-black border-t-2 border-dashed my-4" />
		{/if}
	{/each}
</div>

{#if filteredPosts.length == 0 && !isSaving}
	<div class="mt-4 italic">Nothing {$page.url.pathname.substring(1)} yet...</div>
{/if}

<svelte:head>
	<title>{capitalizeFirstLetter($page.url.pathname.substring(1))} - Lucentsave</title>
</svelte:head>
