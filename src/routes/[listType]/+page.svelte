<script lang="ts">
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { filterPosts } from './util';
	import { markAsRead, like } from '$lib/postActions';
	import type { Post } from './+page.server';
	import { PUBLIC_APPLICATION_URL, PUBLIC_BACKEND_API_URL } from '$env/static/public';

	import { Readability, isProbablyReaderable } from '@mozilla/readability';
	import { onMount } from 'svelte';
	import { isSignedIn, posts } from '../../stores';
	import { postsLoaded } from '../../stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';

	// export let data: PageData;

	if (browser && !$postsLoaded) {
		const fetchPosts = async () => {
			const response = await fetch(PUBLIC_BACKEND_API_URL + 'getAllUserPosts', {
				credentials: 'include'
			});
			if (response.ok) {
				posts.set(await response.json());
				postsLoaded.set(true);
				isSignedIn.set(true);
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

	async function fetchAndParseURL(event: Event) {
		isSaving = true;
		let urlToSave = url;
		url = '';

		event.preventDefault();

		const response = await fetch(
			PUBLIC_BACKEND_API_URL + `fetchPage?url=${encodeURIComponent(urlToSave)}`,
			{
				credentials: 'include'
			}
		);
		if (!response.ok) {
			savingText = 'Failed to load page';
			return;
		}
		const html = await response.text();

		const parser = new DOMParser();
		let doc = parser.parseFromString(html, 'text/html');

		if (isProbablyReaderable(doc)) {
			let reader = new Readability(doc);
			let article = reader.parse();
			if (article != null) {
				title = article.title;

				// Parse the content as a Document again to be able to manipulate it
				let contentDoc = parser.parseFromString(article.content, 'text/html');

				// Convert relative image URLs to absolute
				let imgs = contentDoc.getElementsByTagName('img');
				for (let img of imgs) {
					let urlObject = new URL(img.src, urlToSave);
					img.src = urlObject.href;
				}

				// Convert relative URLs in anchor tags to absolute
				let links = contentDoc.getElementsByTagName('a');
				for (let link of links) {
					if (!link.href.startsWith('#')) {
						// skip navigation links
						let urlObject = new URL(link.href, urlToSave);
						link.href = urlObject.href;
					}
				}

				body = contentDoc.body.innerHTML;

				// console.log('title is: ', title);
				// console.log('body is: ', body);
				await sendPost(urlToSave);
			} else {
				savingText = 'Failed to parse article';
			}
		} else {
			savingText = 'Failed to parse article';
		}
	}

	const sendPost = async (url: string): Promise<void> => {
		const response = await fetch(PUBLIC_BACKEND_API_URL + 'createPost', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ url, title, body }),
			credentials: 'include'
		});
		if (response.ok) {
			posts.set(await response.json());
			isSaving = false;
		} else {
			savingText = await response.text();
		}
	};

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
</script>

{#if $page.url.pathname.startsWith('/saved')}
	<form on:submit={fetchAndParseURL} class="mt-5 flex items-center space-x-2">
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

{#if filteredPosts.length == 0}
	<div class="mt-4 italic">Nothing {$page.url.pathname.substring(1)} yet...</div>
{/if}
