const fenceOpener = /^\s*(```|~~~)/;
const headingLine = /^#{1,6}\s/;
const horizontalRule = /^([-*_])(\s*\1){2,}\s*$/;
const listLine = /^\s*([-*+]|\d+[.)])\s/;

function splitBlocks(markdown: string) {
	const blocks: string[][] = [];
	let current: string[] = [];
	let fence: string | undefined;

	const flush = () => {
		if (current.length) {
			blocks.push(current);
		}
		current = [];
	};

	for (const line of markdown.replace(/\r\n?/g, "\n").split("\n")) {
		if (fence) {
			current.push(line);
			if (line.trim().startsWith(fence)) {
				fence = undefined;
				flush();
			}
			continue;
		}

		const opener = line.match(fenceOpener);
		if (opener) {
			flush();
			fence = opener[1];
			current.push(line);
			continue;
		}

		if (!line.trim()) {
			flush();
			continue;
		}

		current.push(line);
	}

	flush();
	return blocks;
}

function isSkippedBlock(lines: string[]) {
	const first = lines[0].trim();

	if (fenceOpener.test(first)) {
		return true;
	}
	if (headingLine.test(first)) {
		return true;
	}
	if (horizontalRule.test(first)) {
		return true;
	}
	if (first.startsWith("<")) {
		return true;
	}
	if (lines.every((line) => line.trim().startsWith("|"))) {
		return true;
	}
	if (lines.every((line) => listLine.test(line))) {
		return true;
	}
	return false;
}

function stripInlineMarkdown(text: string) {
	let result = text;
	result = result.replace(/!\[[^\]]*\]\([^)]*\)/g, "");
	result = result.replace(/\[([^\]]*)\]\([^)]*\)/g, "$1");
	result = result.replace(/(\*\*|__)(.+?)\1/g, "$2");
	result = result.replace(/~~(.+?)~~/g, "$1");
	result = result.replace(/(\*|_)(.+?)\1/g, "$2");
	result = result.replace(/`([^`]*)`/g, "$1");
	return result;
}

/**
 * Plain-text summary of a Markdown document: the first paragraph block, inline
 * formatting stripped, whitespace collapsed, cut at a word boundary to `limit`
 * characters with a trailing "…" when cut. Undefined when no paragraph exists.
 */
export function markdownSummary(markdown: string, limit = 280): string | undefined {
	const paragraph = splitBlocks(markdown).find((lines) => !isSkippedBlock(lines));
	if (!paragraph) {
		return undefined;
	}

	const unquoted = paragraph.map((line) => line.replace(/^\s*>\s?/, ""));
	const text = stripInlineMarkdown(unquoted.join(" ")).replace(/\s+/g, " ").trim();
	if (!text) {
		return undefined;
	}

	if (text.length <= limit) {
		return text;
	}

	const cutAt = text.lastIndexOf(" ", limit - 1);
	const end = cutAt > 0 ? cutAt : limit - 1;
	return `${text.slice(0, end).trimEnd()}…`;
}
