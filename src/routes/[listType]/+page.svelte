<script lang="ts">
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { filterPosts } from './util';
	import { markAsRead, like } from '$lib/postActions';
	import type { Post } from './+page';

	import DOMPurify from 'dompurify';

	import { Readability, isProbablyReaderable } from '@mozilla/readability';

	export let data: PageData;

	let url: string = '';
	let title: string = '';
	let body: string = '';

	$: isUrlValid = /^http(s)?:\/\/[^\s$.?#].[^\s]*$/.test(url);

	async function fetchAndParseURL(event: Event) {
		event.preventDefault();

		const response = await fetch(
			`http://localhost:8080/api/fetchPage?url=${encodeURIComponent(url)}`,
			{
				credentials: 'include'
			}
		);
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
					let urlObject = new URL(img.src);
					let postUrlObject = new URL(url);

					// Check if the image source has the localhost origin
					if (urlObject.origin === 'http://localhost:5173') {
						// Replace the origin in the image source with the origin of the post URL
						img.src = postUrlObject.origin + urlObject.pathname + urlObject.search + urlObject.hash;
					}
				}

				body = contentDoc.body.innerHTML;
				body = DOMPurify.sanitize(body);

				// console.log('title is: ', title);
				// console.log('body is: ', body);
				await sendPost();
			}
		}
	}

	const sendPost = async (): Promise<void> => {
		data.posts = await (
			await fetch('http://localhost:8080/api/createPost', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ url, title, body }),
				credentials: 'include'
			})
		).json();

		data.posts = filterPosts(data.posts, $page.url.pathname.substring(1));
	};

	const updatePost = async (post: Post) => {
		const postIndex = data.posts.findIndex((p) => p.id === post.id);

		if (postIndex !== -1) {
			data.posts[postIndex] = post;
		}

		data.posts = filterPosts(data.posts, $page.url.pathname.substring(1));
	};

	const likePostAndUpdate = async (post: Post) => {
		const udpatedPost = await like(post);
		updatePost(udpatedPost);
	};

	const markPostReadAndUpdate = async (post: Post) => {
		const udpatedPost = await markAsRead(post);
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
			class="py-1 px-2 border-2 border-black cursor-pointer {isUrlValid
				? 'bg-black text-white hover:bg-gray-700'
				: 'bg-gray-700 text-white cursor-not-allowed'}"
			disabled={!isUrlValid}
		/>
	</form>
{/if}

<div class="mt-4">
	{#each data.posts as post (post.id)}
		<div class="flex justify-between items-center">
			<a href={`/post/${post.id}`} class="hover:text-gray-500">
				<div class="text-2xl font-bold block">{post.title}</div>
				<div class="text-sm block">{getHostname(post.url)}</div>
			</a>
			<div>
				{#if !$page.url.pathname.startsWith('/saved')}
					<span
						role="button"
						tabindex="0"
						on:click={() => likePostAndUpdate(post)}
						on:keydown={(e) => e.key === 'Enter' && likePostAndUpdate(post)}
						class="text-black px-2 py-1 text-xl cursor-pointer hover:text-gray-500"
						>{post.isLiked ? '★' : '☆'}</span
					>
				{/if}
			</div>
		</div>
		{#if post.id !== data.posts[data.posts.length - 1].id}
			<hr class="border-black border-t-2 border-dashed my-4" />
		{/if}
	{/each}
</div>

{#if data.posts.length == 0}
	<div class="mt-4 italic">Nothing {$page.url.pathname.substring(1)} yet...</div>
{/if}
