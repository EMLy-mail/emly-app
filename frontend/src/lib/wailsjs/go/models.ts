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

export namespace gpu {
	
	export class GraphicsCard {
	    address: string;
	    index: number;
	    pci?: pci.Device;
	    node?: topology.Node;
	
	    static createFrom(source: any = {}) {
	        return new GraphicsCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.index = source["index"];
	        this.pci = this.convertValues(source["pci"], pci.Device);
	        this.node = this.convertValues(source["node"], topology.Node);
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
	    cards: GraphicsCard[];
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cards = this.convertValues(source["cards"], GraphicsCard);
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
	export class Area {
	    total_physical_bytes: number;
	    total_usable_bytes: number;
	    supported_page_sizes: number[];
	    default_huge_page_size: number;
	    total_huge_page_bytes: number;
	    huge_page_amounts_by_size: Record<number, HugePageAmounts>;
	    modules: Module[];
	
	    static createFrom(source: any = {}) {
	        return new Area(source);
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
	export class Cache {
	    level: number;
	    type: number;
	    size_bytes: number;
	    logical_processors: number[];
	
	    static createFrom(source: any = {}) {
	        return new Cache(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.type = source["type"];
	        this.size_bytes = source["size_bytes"];
	        this.logical_processors = source["logical_processors"];
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

export namespace pci {
	
	export class Device {
	    address: string;
	    parent_address: string;
	    vendor?: types.Vendor;
	    product?: types.Product;
	    revision: string;
	    subsystem?: types.Product;
	    class?: types.Class;
	    subclass?: types.Subclass;
	    programming_interface?: types.ProgrammingInterface;
	    node?: topology.Node;
	    driver: string;
	    iommu_group: string;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.parent_address = source["parent_address"];
	        this.vendor = this.convertValues(source["vendor"], types.Vendor);
	        this.product = this.convertValues(source["product"], types.Product);
	        this.revision = source["revision"];
	        this.subsystem = this.convertValues(source["subsystem"], types.Product);
	        this.class = this.convertValues(source["class"], types.Class);
	        this.subclass = this.convertValues(source["subclass"], types.Subclass);
	        this.programming_interface = this.convertValues(source["programming_interface"], types.ProgrammingInterface);
	        this.node = this.convertValues(source["node"], topology.Node);
	        this.driver = source["driver"];
	        this.iommu_group = source["iommu_group"];
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

export namespace topology {
	
	export class Node {
	    id: number;
	    cores: cpu.ProcessorCore[];
	    caches: memory.Cache[];
	    distances: number[];
	    memory?: memory.Area;
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.cores = this.convertValues(source["cores"], cpu.ProcessorCore);
	        this.caches = this.convertValues(source["caches"], memory.Cache);
	        this.distances = source["distances"];
	        this.memory = this.convertValues(source["memory"], memory.Area);
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

export namespace types {
	
	export class ProgrammingInterface {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ProgrammingInterface(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class Subclass {
	    id: string;
	    name: string;
	    programming_interfaces: ProgrammingInterface[];
	
	    static createFrom(source: any = {}) {
	        return new Subclass(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.programming_interfaces = this.convertValues(source["programming_interfaces"], ProgrammingInterface);
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
	export class Class {
	    id: string;
	    name: string;
	    subclasses: Subclass[];
	
	    static createFrom(source: any = {}) {
	        return new Class(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.subclasses = this.convertValues(source["subclasses"], Subclass);
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
	export class Product {
	    vendor_id: string;
	    id: string;
	    name: string;
	    subsystems: Product[];
	
	    static createFrom(source: any = {}) {
	        return new Product(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vendor_id = source["vendor_id"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.subsystems = this.convertValues(source["subsystems"], Product);
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
	
	
	export class Vendor {
	    id: string;
	    name: string;
	    products: Product[];
	
	    static createFrom(source: any = {}) {
	        return new Vendor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.products = this.convertValues(source["products"], Product);
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
	
	    static createFrom(source: any = {}) {
	        return new EMLyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SDKDecoderSemver = source["SDKDecoderSemver"];
	        this.SDKDecoderReleaseChannel = source["SDKDecoderReleaseChannel"];
	        this.GUISemver = source["GUISemver"];
	        this.GUIReleaseChannel = source["GUIReleaseChannel"];
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
	    ExternalIP: string;
	    CPU: cpu.Info;
	    RAM: memory.Info;
	    GPU: gpu.Info;
	
	    static createFrom(source: any = {}) {
	        return new MachineInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hostname = source["Hostname"];
	        this.OS = source["OS"];
	        this.Version = source["Version"];
	        this.HWID = source["HWID"];
	        this.ExternalIP = source["ExternalIP"];
	        this.CPU = this.convertValues(source["CPU"], cpu.Info);
	        this.RAM = this.convertValues(source["RAM"], memory.Info);
	        this.GPU = this.convertValues(source["GPU"], gpu.Info);
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

