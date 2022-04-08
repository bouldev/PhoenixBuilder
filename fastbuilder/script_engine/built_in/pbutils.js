
// This script represents simple APIs can be done by Javascript.
// And do initialization works

engine.message=(msg)=> {
	printf("%s\n",msg);
};

const console={
	log: function (){
		let args=[...arguments];
		args[0]+="\n";
		printf(...args);
	}
};

class SimpleStorage {
	constructor(path) {
		if(fs.containerPath=="") {
			throw new Error("No container created");
		}
		this.preferencesFn=consts.bundle.fromRequire?`pref_${consts.bundle.currentScript}.json`:"preferences.json";
		if(!fs.exists(this.preferencesFn)) {
			fs.writeFile(this.preferencesFn, "{}");
		}else{
			let prefsContent=fs.readFile(this.preferencesFn);
			let prefs;
			try {
				prefs=JSON.parse(prefsContent);
			}catch(err) {
				throw new Error(sprintf("Corrupted preferences file %s: %s",this.preferencesFn,err.message));
			}
			this.prefs=prefs;
		}
	}
	
	Set(key, value) {
		this.prefs[key]=value;
		fs.writeFile(this.preferencesFn, JSON.stringify(this.prefs, null, "\t"));
	}
	
	Get(key) {
		return this.prefs[key];
	}
	
	Has(key) {
		return this.prefs.hasOwnProperty(key);
	}
	
	Delete(key) {
		delete this.prefs[key];
		fs.writeFile(this.preferencesFn, JSON.stringify(this.prefs, null, "\t"));
		return true;
	}
}

let Storage;

function require(module_name) {
	if(!module.require) {
		throw new Error("require() outside a bundle is not permitted.");
	}
	return module.require(module_name);
}

if(typeof(consts.bundle)=="object") {
	let manifest=JSON.parse(consts.bundle.manifest);
	if(!consts.bundle.fromRequire) {
		consts.bundle.currentScript=manifest.entrypoint;
	}
	if(manifest.no_container) {
		delete fs;
	}else{
		if(!consts.bundle.fromRequire) {
			fs.requireContainer(consts.bundle.identifier);
		}
		Storage=new SimpleStorage();
	}
	delete engine.setName;
	if(!consts.bundle.fromRequire) {
		engine.setNameInternal(consts.bundle.name);
	}
	module.require=(name)=>{
		if(!consts.bundle.content.hasOwnProperty(name)) {
			throw new Error(sprintf("Cannot find module '%s'",name));
		}
		let c=consts.bundle.content[name];
		if(typeof(c)=="string") {
			try {
				return JSON.parse(c);
			}catch(e) {
				// Do as what Node.JS does.
				return {};
			}
		}
		if(!c.hasOwnProperty("run")) {
			throw new Error(sprintf("Illegal module '%s'",name));
		}
		return c.run();
	};
	if(!consts.bundle.fromRequire) {
		printf("Loaded Script Bundle %s@%s\n",consts.bundle.name,consts.bundle.version);
		if(manifest.custom_welcome_text) {
			printf("%s\n",manifest.custom_welcome_text);
		}
	}
	delete engine.setNameInternal;
}else{
	delete engine.setNameInternal;
}

