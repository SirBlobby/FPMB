import { setAccessToken } from '$lib/api/client';
import { auth as authApi, users as usersApi } from '$lib/api';
import type { User } from '$lib/types/api';

function createAuthStore() {
	let user = $state<User | null>(null);
	let loading = $state(true);

	async function init() {
		const token =
			typeof localStorage !== 'undefined' ? localStorage.getItem('access_token') : null;
		if (!token) {
			loading = false;
			return;
		}
		try {
			user = await usersApi.me();
		} catch {
			user = null;
		} finally {
			loading = false;
		}
	}

	async function login(email: string, password: string) {
		const res = await authApi.login(email, password);
		setAccessToken(res.access_token);
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem('refresh_token', res.refresh_token);
			localStorage.setItem('user_id', res.user.id);
		}
		user = res.user;
	}

	async function register(name: string, email: string, password: string) {
		const res = await authApi.register(name, email, password);
		setAccessToken(res.access_token);
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem('refresh_token', res.refresh_token);
			localStorage.setItem('user_id', res.user.id);
		}
		user = res.user;
	}

	async function logout() {
		try {
			await authApi.logout();
		} catch {
		}
		setAccessToken(null);
		if (typeof localStorage !== 'undefined') {
			localStorage.removeItem('refresh_token');
			localStorage.removeItem('user_id');
		}
		user = null;
	}

	function setUser(u: User) {
		user = u;
	}

	return {
		get user() {
			return user;
		},
		get loading() {
			return loading;
		},
		init,
		login,
		register,
		logout,
		setUser
	};
}

export const authStore = createAuthStore();
