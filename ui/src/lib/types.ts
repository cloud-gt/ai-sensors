export interface Command {
	id: string;
	name: string;
	command: string;
	work_dir: string;
}

export interface CommandListResponse {
	commands: Command[];
}

export interface StatusResponse {
	status: 'running' | 'stopped' | 'not_started';
}

export interface OutputResponse {
	lines: string[];
}

export interface StartResponse {
	started: boolean;
}

export interface ErrorResponse {
	error: string;
}
