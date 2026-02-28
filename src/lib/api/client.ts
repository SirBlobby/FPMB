const BASE = '/api';

let accessToken: string | null =
	typeof localStorage !== 'undefined' ? localStorage.getItem('access_token') : null;

export function getAccessToken() {
	return accessToken;
}

export function setAccessToken(token: string | null) {
	accessToken = token;
	if (typeof localStorage !== 'undefined') {
		if (token) {
			localStorage.setItem('access_token', token);
		} else {
			localStorage.removeItem('access_token');
		}
	}
}

async function refreshAccessToken(): Promise<string | null> {
	const refreshToken =
		typeof localStorage !== 'undefined' ? localStorage.getItem('refresh_token') : null;
	if (!refreshToken) return null;

	const res = await fetch(`${BASE}/auth/refresh`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ refresh_token: refreshToken })
	});

	if (!res.ok) {
		setAccessToken(null);
		if (typeof localStorage !== 'undefined') localStorage.removeItem('refresh_token');
		return null;
	}

	const data = await res.json();
	setAccessToken(data.access_token);
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem('refresh_token', data.refresh_token);
	}
	return data.access_token;
}

export async function apiFetch<T>(
	path: string,
	options: RequestInit = {},
	retry = true
): Promise<T> {
	const token = accessToken;
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(options.headers as Record<string, string>)
	};
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(`${BASE}${path}`, { ...options, headers });

	if (res.status === 401 && retry) {
		const newToken = await refreshAccessToken();
		if (newToken) return apiFetch<T>(path, options, false);
		throw new Error('Unauthorized');
	}

	if (!res.ok) {
		const body = await res.json().catch(() => ({}));
		throw new Error(body.error || `HTTP ${res.status}`);
	}

	if (res.status === 204) return undefined as T;
	return res.json();
}

export async function apiFetchFormData<T>(
	path: string,
	formData: FormData,
	retry = true
): Promise<T> {
	const token = accessToken;
	const headers: Record<string, string> = {};
	if (token) headers['Authorization'] = `Bearer ${token}`;

	const res = await fetch(`${BASE}${path}`, { method: 'POST', body: formData, headers });

	if (res.status === 401 && retry) {
		const newToken = await refreshAccessToken();
		if (newToken) return apiFetchFormData<T>(path, formData, false);
		throw new Error('Unauthorized');
	}

	if (!res.ok) {
		const body = await res.json().catch(() => ({}));
		throw new Error(body.error || `HTTP ${res.status}`);
	}

	if (res.status === 204) return undefined as T;
	return res.json();
}
