engine.message("hello from part2.js");
// External values are not accessible, exporting a function requiring them if you're
// intrested in them
let impv;

module.exports={
	doit: (imp)=>{
		impv=imp;
	},
	get: ()=> {
		return impv;
	}
}