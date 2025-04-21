package gosolc

const wrapperScript = `
var Module = {
	locateFile: function(path) { return path; },
	print: function(x) { console.log(x); },
	printErr: function(x) { console.error(x); },
	onRuntimeInitialized: function() {
		if (typeof Module.cwrap === 'function' && typeof Module._solidity_compile === 'function') {
			solc.compile = Module.cwrap('solidity_compile', 'string', ['string', 'string']);
		} else {
			throw new Error('solidity_compile or cwrap not available');
		}
	}
};
var console = { log: function(x) {}, error: function(x) {} }; // Mock console
var solc = {};
%s
// Manually trigger initialization
if (Module.asm && Module.asm.compile) {
	Module.onRuntimeInitialized();
} else if (Module._solidity_compile) {
	Module.onRuntimeInitialized();
}
`
