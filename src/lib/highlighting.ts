function highlightNode (node: Node, highlightID: string) {
	const span = document.createElement('span')
	// span.style.backgroundColor = 'yellow'
	span.classList.add('bg-yellow-200')
	span.style.cursor = 'pointer'
	span.id = highlightID
	span.dataset.highlightId = highlightID
	node.parentNode?.replaceChild(span, node)
	span.appendChild(node)
}

export function highlightRange (range: Range, highlightID: string) {
	const completelyHighlighted = getCompletelyHighlightedNodes(range)
	const edgeHighlighted = getEdgeHighlightedNodes(range)

	for (const node of completelyHighlighted) {
		highlightNode(node, highlightID)
	}

	for (const node of edgeHighlighted) {
		highlightNode(node, highlightID)
	}
}

export function isNodeInRange (range: Range, node: Node): boolean {
	const nodeRange = document.createRange()
	nodeRange.selectNode(node)

	const START_TO_START = 0
	const END_TO_END = 2

	return (
		range.compareBoundaryPoints(START_TO_START, nodeRange) <= 0 &&
		range.compareBoundaryPoints(END_TO_END, nodeRange) >= 0
	)
}

function getCompletelyHighlightedNodes (range: Range): Node[] {
	const iter = document.createNodeIterator(range.commonAncestorContainer, NodeFilter.SHOW_TEXT)
	let currentNode = iter.nextNode()

	let res: Node[] = []

	while (currentNode) {
		if (range.intersectsNode(currentNode) && isNodeInRange(range, currentNode)) {
			// console.log(currentNode.textContent);
			res.push(currentNode)
		}
		currentNode = iter.nextNode()
	}

	return res
}

function getEdgeHighlightedNodes (range: Range): Node[] {
	const { startContainer, startOffset, endContainer, endOffset } = range
	let edgeNodes: Node[] = []

	if (startContainer === endContainer && startContainer.nodeType === Node.TEXT_NODE) {
		;(startContainer as Text).splitText(endOffset)
		let newStartNode = (startContainer as Text).splitText(startOffset)
		edgeNodes.push(newStartNode)
	} else {
		if (startContainer.nodeType === Node.TEXT_NODE) {
			let newStartNode = (startContainer as Text).splitText(startOffset)
			edgeNodes.push(newStartNode)
		}

		if (endContainer.nodeType === Node.TEXT_NODE) {
			;(endContainer as Text).splitText(endOffset)
			edgeNodes.push(endContainer)
		}
	}

	return edgeNodes
}
