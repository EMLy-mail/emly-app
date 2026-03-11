import type { utils } from '$lib/wailsjs/go/models';

class SystemInfoStore {
	data = $state<utils.MachineInfo | null>(null);

	setData(info: utils.MachineInfo) {
		this.data = info;
	}

	clear() {
		this.data = null;
	}
}

export const systemInfoStore = new SystemInfoStore();
