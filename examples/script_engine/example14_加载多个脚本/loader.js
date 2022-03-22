// 本脚本路径
const thisScriptPath=consts.script_path;
// 按 '/' 分割
const pathStrSplit = thisScriptPath.split('/');
// 丢掉最后的路径
const fileName = pathStrSplit.pop();
// 获得文件夹名字
const thisDirName = pathStrSplit.join('/');
// 获得绝对路径
const thisDirAbsName = consts.fb_dir+'/'+thisDirName;
engine.message(thisDirAbsName);

// 请求读取文件的权限
let isReqSuccess=fs.requestFilePermission(thisDirAbsName,"需要访问这个文件夹以加载文件剩余部分");
if (!isReqSuccess) {
	// 如果你的脚本必须要文件权限才能正常工作
	// 你可以使用该函数主动崩溃脚本
	engine.message("没有获得权限!");
	engine.crash("必须这个文件夹的权限才能工作");
}

// 需要加载的脚本文件
const scripts=[
	"part1.js",
	"part2.js"
];

let globalVar;

for(let script of scripts){
	engine.message("loading: "+thisDirAbsName+"/"+script);
	let data=fs.readFile(thisDirAbsName+"/"+script);
	eval(data);
}

// engine.waitConnectionSync();
