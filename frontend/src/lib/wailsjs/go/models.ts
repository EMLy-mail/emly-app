export namespace cpu {
	
	export class ProcessorCore {
	    id: number;
	    total_hardware_threads: number;
	    total_threads: number;
	    logical_processors: number[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessorCore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.total_hardware_threads = source["total_hardware_threads"];
	        this.total_threads = source["total_threads"];
	        this.logical_processors = source["logical_processors"];
	    }
	}
	export class Processor {
	    id: number;
	    total_cores: number;
	    total_hardware_threads: number;
	    total_threads: number;
	    vendor: string;
	    model: string;
	    capabilities: string[];
	    cores: ProcessorCore[];
	
	    static createFrom(source: any = {}) {
	        return new Processor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.total_cores = source["total_cores"];
	        this.total_hardware_threads = source["total_hardware_threads"];
	        this.total_threads = source["total_threads"];
	        this.vendor = source["vendor"];
	        this.model = source["model"];
	        this.capabilities = source["capabilities"];
	        this.cores = this.convertValues(source["cores"], ProcessorCore);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Info {
	    total_cores: number;
	    total_hardware_threads: number;
	    total_threads: number;
	    processors: Processor[];
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_cores = source["total_cores"];
	        this.total_hardware_threads = source["total_hardware_threads"];
	        this.total_threads = source["total_threads"];
	        this.processors = this.convertValues(source["processors"], Processor);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace internal {
	
	export class EmailAttachment {
	    filename: string;
	    contentType: string;
	    data: number[];
	
	    static createFrom(source: any = {}) {
	        return new EmailAttachment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.contentType = source["contentType"];
	        this.data = source["data"];
	    }
	}
	export class EmailData {
	    from: string;
	    to: string[];
	    cc: string[];
	    bcc: string[];
	    subject: string;
	    body: string;
	    attachments: EmailAttachment[];
	    isPec: boolean;
	    hasInnerEmail: boolean;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new EmailData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.cc = source["cc"];
	        this.bcc = source["bcc"];
	        this.subject = source["subject"];
	        this.body = source["body"];
	        this.attachments = this.convertValues(source["attachments"], EmailAttachment);
	        this.isPec = source["isPec"];
	        this.hasInnerEmail = source["hasInnerEmail"];
	        this.date = source["date"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class BugReportInput {
	    name: string;
	    email: string;
	    description: string;
	    screenshotData: string;
	    localStorageData: string;
	    configData: string;
	
	    static createFrom(source: any = {}) {
	        return new BugReportInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.email = source["email"];
	        this.description = source["description"];
	        this.screenshotData = source["screenshotData"];
	        this.localStorageData = source["localStorageData"];
	        this.configData = source["configData"];
	    }
	}
	export class BugReportResult {
	    folderPath: string;
	    screenshotPath: string;
	    mailFilePath: string;
	
	    static createFrom(source: any = {}) {
	        return new BugReportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folderPath = source["folderPath"];
	        this.screenshotPath = source["screenshotPath"];
	        this.mailFilePath = source["mailFilePath"];
	    }
	}
	export class ImageViewerData {
	    data: string;
	    filename: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageViewerData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.filename = source["filename"];
	    }
	}
	export class PDFViewerData {
	    data: string;
	    filename: string;
	
	    static createFrom(source: any = {}) {
	        return new PDFViewerData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.filename = source["filename"];
	    }
	}
	export class ScreenshotResult {
	    data: string;
	    width: number;
	    height: number;
	    filename: string;
	
	    static createFrom(source: any = {}) {
	        return new ScreenshotResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.data = source["data"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.filename = source["filename"];
	    }
	}
	export class SubmitBugReportResult {
	    zipPath: string;
	    folderPath: string;
	    uploaded: boolean;
	    reportId: number;
	    uploadError: string;
	
	    static createFrom(source: any = {}) {
	        return new SubmitBugReportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zipPath = source["zipPath"];
	        this.folderPath = source["folderPath"];
	        this.uploaded = source["uploaded"];
	        this.reportId = source["reportId"];
	        this.uploadError = source["uploadError"];
	    }
	}
	export class UpdateStatus {
	    currentVersion: string;
	    availableVersion: string;
	    updateAvailable: boolean;
	    checking: boolean;
	    downloading: boolean;
	    downloadProgress: number;
	    ready: boolean;
	    installerPath: string;
	    errorMessage: string;
	    releaseNotes?: string;
	    severityType?: string;
	    lastCheckTime: string;
	    channel?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.availableVersion = source["availableVersion"];
	        this.updateAvailable = source["updateAvailable"];
	        this.checking = source["checking"];
	        this.downloading = source["downloading"];
	        this.downloadProgress = source["downloadProgress"];
	        this.ready = source["ready"];
	        this.installerPath = source["installerPath"];
	        this.errorMessage = source["errorMessage"];
	        this.releaseNotes = source["releaseNotes"];
	        this.severityType = source["severityType"];
	        this.lastCheckTime = source["lastCheckTime"];
	        this.channel = source["channel"];
	    }
	}
	export class ViewerData {
	    imageData?: ImageViewerData;
	    pdfData?: PDFViewerData;
	
	    static createFrom(source: any = {}) {
	        return new ViewerData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imageData = this.convertValues(source["imageData"], ImageViewerData);
	        this.pdfData = this.convertValues(source["pdfData"], PDFViewerData);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace memory {
	
	export class Module {
	    label: string;
	    location: string;
	    serial_number: string;
	    size_bytes: number;
	    vendor: string;
	
	    static createFrom(source: any = {}) {
	        return new Module(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.location = source["location"];
	        this.serial_number = source["serial_number"];
	        this.size_bytes = source["size_bytes"];
	        this.vendor = source["vendor"];
	    }
	}
	export class HugePageAmounts {
	    total: number;
	    free: number;
	    surplus: number;
	    reserved: number;
	
	    static createFrom(source: any = {}) {
	        return new HugePageAmounts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.free = source["free"];
	        this.surplus = source["surplus"];
	        this.reserved = source["reserved"];
	    }
	}
	export class Info {
	    total_physical_bytes: number;
	    total_usable_bytes: number;
	    supported_page_sizes: number[];
	    default_huge_page_size: number;
	    total_huge_page_bytes: number;
	    huge_page_amounts_by_size: Record<number, HugePageAmounts>;
	    modules: Module[];
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_physical_bytes = source["total_physical_bytes"];
	        this.total_usable_bytes = source["total_usable_bytes"];
	        this.supported_page_sizes = source["supported_page_sizes"];
	        this.default_huge_page_size = source["default_huge_page_size"];
	        this.total_huge_page_bytes = source["total_huge_page_bytes"];
	        this.huge_page_amounts_by_size = this.convertValues(source["huge_page_amounts_by_size"], HugePageAmounts, true);
	        this.modules = this.convertValues(source["modules"], Module);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace utils {
	
	export class EMLyConfig {
	    SDKDecoderSemver: string;
	    SDKDecoderReleaseChannel: string;
	    GUISemver: string;
	    GUIReleaseChannel: string;
	    Language: string;
	    UpdateCheckEnabled: string;
	    UpdatePath: string;
	    UpdateAutoCheck: string;
	    BugReportAPIURL: string;
	    BugReportAPIKey: string;
	    LogLevel: string;
	
	    static createFrom(source: any = {}) {
	        return new EMLyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SDKDecoderSemver = source["SDKDecoderSemver"];
	        this.SDKDecoderReleaseChannel = source["SDKDecoderReleaseChannel"];
	        this.GUISemver = source["GUISemver"];
	        this.GUIReleaseChannel = source["GUIReleaseChannel"];
	        this.Language = source["Language"];
	        this.UpdateCheckEnabled = source["UpdateCheckEnabled"];
	        this.UpdatePath = source["UpdatePath"];
	        this.UpdateAutoCheck = source["UpdateAutoCheck"];
	        this.BugReportAPIURL = source["BugReportAPIURL"];
	        this.BugReportAPIKey = source["BugReportAPIKey"];
	        this.LogLevel = source["LogLevel"];
	    }
	}
	export class Config {
	    EMLy: EMLyConfig;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.EMLy = this.convertValues(source["EMLy"], EMLyConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MachineInfo {
	    Hostname: string;
	    OS: string;
	    Version: string;
	    HWID: string;
	    CPU: cpu.Info;
	    RAM: memory.Info;
	
	    static createFrom(source: any = {}) {
	        return new MachineInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hostname = source["Hostname"];
	        this.OS = source["OS"];
	        this.Version = source["Version"];
	        this.HWID = source["HWID"];
	        this.CPU = this.convertValues(source["CPU"], cpu.Info);
	        this.RAM = this.convertValues(source["RAM"], memory.Info);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

