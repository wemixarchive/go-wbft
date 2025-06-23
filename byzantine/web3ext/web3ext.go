package web3ext

const Byzantine_JS = `
web3._extend({
	property: 'byzantine',
	methods: [
		new web3._extend.Method({
			name: 'stopByzantineTests',
			call: 'byzantine_stopByzantineTests',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'silentMessage',
			call: 'byzantine_silentMessage',
			params: 1,
			inputFormatter: [null]
		}),
			new web3._extend.Method({
			name: 'sendTamperedMessage',
			call: 'byzantine_sendTamperedMessage',
			params: 1,
			inputFormatter: [null]
		}),
			new web3._extend.Method({
			name: 'sendFakeMessage',
			call: 'byzantine_sendFakeMessage',
			params: 1,
			inputFormatter: [null]
		}),
			new web3._extend.Method({
			name: 'sendOmitMessage',
			call: 'byzantine_sendOmitMessage',
			params: 1,
			inputFormatter: [null]
		}),
			new web3._extend.Method({
			name: 'sendRoleSpoofedMessage',
			call: 'byzantine_sendRoleSpoofedMessage',
			params: 1,
			inputFormatter: [null]
		}),
			new web3._extend.Method({
			name: 'sendReplayMessage',
			call: 'byzantine_sendReplayMessage',
			params: 1,
			inputFormatter: [null]
		}),
			new web3._extend.Method({
			name: 'upgradeGovContract',
			call: 'byzantine_upgradeGovContract',
			params: 0
		}),
	],
	properties:
	[
		new web3._extend.Property({
			name: 'byzantineTests',
			getter: 'byzantine_byzantineTests'
		}),
	]
});
`
