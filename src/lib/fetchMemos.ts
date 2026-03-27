// Memos API data abstraction layer (v2 API — /api/v1/memos)
// PUBLIC_MEMOS_API_BASE is Astro's convention for browser-accessible env vars.
// process.env fallback covers build-time Node.js contexts (e.g. astro.config.mjs).
export const MEMOS_API_BASE =
	import.meta.env.PUBLIC_MEMOS_API_BASE ||
	(typeof process !== "undefined" && process.env.PUBLIC_MEMOS_API_BASE) ||
	"http://127.0.0.1:5230";

const MEMOS_API_URL = `${MEMOS_API_BASE}/api/v1/memos`;

/**
 * Memos v2 API response shape
 */
export interface MemoV2 {
	name: string;
	state: string;
	creator: string;
	createTime: string;
	updateTime: string;
	displayTime: string;
	content: string;
	visibility: string;
	tags: string[];
	pinned: boolean;
	snippet: string;
	attachments: MemoAttachment[];
	property: {
		hasLink: boolean;
		hasTaskList: boolean;
		hasCode: boolean;
		hasIncompleteTasks: boolean;
	};
	location?: {
		placeholder: string;
		latitude: number;
		longitude: number;
	};
}

export interface MemoAttachment {
	name: string;
	createTime: string;
	filename: string;
	externalLink: string;
	type: string;
	size: string;
	memo: string;
}

interface MemosResponse {
	memos: MemoV2[];
	nextPageToken: string;
}

/**
 * Fetch the latest N public memos.
 * Returns an empty array on failure for graceful degradation.
 */
export async function fetchMemos(limit: number = 5): Promise<MemoV2[]> {
	try {
		const res = await fetch(
			`${MEMOS_API_URL}?filter=visibility=="PUBLIC"&pageSize=${limit}`
		);
		if (!res.ok) {
			console.error(`[fetchMemos] API responded with ${res.status}`);
			return [];
		}
		const data: MemosResponse = await res.json();
		return data.memos ?? [];
	} catch (err) {
		console.error("[fetchMemos] Network error:", err);
		return [];
	}
}

/**
 * Fetch ALL public memos (for paginated archive pages).
 * Handles pagination via nextPageToken.
 */
export async function fetchAllMemos(): Promise<MemoV2[]> {
	const all: MemoV2[] = [];
	let pageToken = "";

	try {
		do {
			const url = pageToken
				? `${MEMOS_API_URL}?filter=visibility=="PUBLIC"&pageSize=50&pageToken=${pageToken}`
				: `${MEMOS_API_URL}?filter=visibility=="PUBLIC"&pageSize=50`;

			const res = await fetch(url);
			if (!res.ok) {
				console.error(`[fetchAllMemos] API responded with ${res.status}`);
				break;
			}
			const data: MemosResponse = await res.json();
			all.push(...(data.memos ?? []));
			pageToken = data.nextPageToken ?? "";
		} while (pageToken);
	} catch (err) {
		console.error("[fetchAllMemos] Network error:", err);
	}

	return all;
}

/**
 * Format an ISO date string into Chinese date format.
 * e.g. "2026-03-23T15:18:36Z" → "2026年3月23日 23:18"  (converts to local time)
 */
export function formatMemoDate(dateStr: string): string {
	const d = new Date(dateStr);
	const year = d.getFullYear();
	const month = d.getMonth() + 1;
	const day = d.getDate();
	const hours = String(d.getHours()).padStart(2, "0");
	const minutes = String(d.getMinutes()).padStart(2, "0");
	return `${year}年${month}月${day}日 ${hours}:${minutes}`;
}

/**
 * Format an ISO date string into a short time string.
 * e.g. [23:18]
 */
export function formatMemoTime(dateStr: string): string {
	const d = new Date(dateStr);
	const hours = String(d.getHours()).padStart(2, "0");
	const minutes = String(d.getMinutes()).padStart(2, "0");
	return `[${hours}:${minutes}]`;
}

/**
 * Build the full URL for a memo attachment.
 * Format: {base}/file/{name}/{encodedFilename}
 */
export function getAttachmentUrl(attachmentName: string, filename: string): string {
	return `${MEMOS_API_BASE}/file/${attachmentName}/${encodeURIComponent(filename)}`;
}
