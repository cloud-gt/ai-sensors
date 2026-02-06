import type {
	Command,
	CommandListResponse,
	StatusResponse,
	OutputResponse,
	StartResponse
} from './types';

const BASE = '/commands';

async function handleResponse<T>(res: Response): Promise<T> {
	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(body.error || `HTTP ${res.status}`);
	}
	return res.json();
}

export async function listCommands(): Promise<Command[]> {
	const data = await handleResponse<CommandListResponse>(await fetch(BASE));
	return data.commands ?? [];
}

export async function getCommand(id: string): Promise<Command> {
	return handleResponse<Command>(await fetch(`${BASE}/${id}`));
}

export async function createCommand(
	name: string,
	command: string,
	workDir: string
): Promise<Command> {
	return handleResponse<Command>(
		await fetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name, command, work_dir: workDir })
		})
	);
}

export async function deleteCommand(id: string): Promise<void> {
	const res = await fetch(`${BASE}/${id}`, { method: 'DELETE' });
	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(body.error || `HTTP ${res.status}`);
	}
}

export async function startCommand(id: string): Promise<boolean> {
	const data = await handleResponse<StartResponse>(
		await fetch(`${BASE}/${id}/start`, { method: 'POST' })
	);
	return data.started;
}

export async function stopCommand(id: string): Promise<void> {
	const res = await fetch(`${BASE}/${id}/stop`, { method: 'POST' });
	if (!res.ok) {
		const body = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(body.error || `HTTP ${res.status}`);
	}
}

export async function getStatus(id: string): Promise<string> {
	const data = await handleResponse<StatusResponse>(
		await fetch(`${BASE}/${id}/status`)
	);
	return data.status;
}

export async function getOutput(id: string, lines?: number): Promise<string[]> {
	const url =
		lines !== undefined ? `${BASE}/${id}/output?lines=${lines}` : `${BASE}/${id}/output`;
	const data = await handleResponse<OutputResponse>(await fetch(url));
	return data.lines ?? [];
}
