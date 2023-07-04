import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import type { Post } from '../../[listType]/+page';

// export const ssr = false;

export const load: PageLoad = async ({ params, fetch }) => {
	const response = await fetch(`http://localhost:8080/api/getPost?id=${params.id}`, {
		credentials: 'include'
	});
	if (response.ok) {
		const post: Post = await response.json();
		return { post };
	}

	throw redirect(307, '/saved');
};
