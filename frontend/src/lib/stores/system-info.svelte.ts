import type { utils } from '$lib/wailsjs/go/models';

class SystemInfoStore {
	data = $state<utils.ExtendedMachineInfo | null>(null);

	setData(info: utils.ExtendedMachineInfo) {
		this.data = info;
	}

	clear() {
		this.data = null;
	}
}

export const systemInfoStore = new SystemInfoStore();
