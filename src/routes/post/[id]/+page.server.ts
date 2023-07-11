import { redirect } from '@sveltejs/kit'
import type { Post } from '../../[listType]/+page.server'
import { PUBLIC_BACKEND_API_URL } from '$env/static/public'
import type { PageServerLoad } from './$types'
import { getNodeFromPath, getPathTo } from '$lib/highlighting'

// ???
export const ssr = false

interface Highlight {
	id: string
	startContainerPath: number[]
	startOffset: number
	endContainerPath: number[]
	endOffset: number
}

interface HighlightRange {
	id: string
	range: Range
}

export const load: PageServerLoad = async ({ params, fetch }) => {
	const response = await fetch(PUBLIC_BACKEND_API_URL + `getPost?id=${params.id}`, {
		credentials: 'include'
	})
	if (response.ok) {
		// Response contains both the post and list of highlight data objects from which we recreate the highlight ranges
		const { post, highlights }: { post: Post; highlights: Highlight[] } = await response.json()

		const highlightRanges: HighlightRange[] = highlights.map((highlight: Highlight) => {
			const range = document.createRange()

			const startNode = getNodeFromPath(highlight.startContainerPath)
			const endNode = getNodeFromPath(highlight.endContainerPath)

			if (
				startNode &&
				endNode &&
				startNode.nodeType === Node.TEXT_NODE &&
				endNode.nodeType === Node.TEXT_NODE
			) {
				range.setStart(startNode, highlight.startOffset)
				range.setEnd(endNode, highlight.endOffset)
			}

			return { id: highlight.id, range }
		})

		return { post, highlightRanges }
	}

	throw redirect(307, '/saved')
}
