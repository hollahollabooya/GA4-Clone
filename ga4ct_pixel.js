window.ga4ct = {
	// Helpful values
	SESSION_EXPIRATION: 1800000, // 30 minutes in milliseconds
	MAX_EPIRATION: 46656000000, // 540 days in milliseconds
	DOMAIN: document.domain.split('.').reverse().splice(0,2).reverse().join('.'), // root domain
	CID_COOKIE_NAME: "_ga4ct_cid",
	SID_COOKIE_NAME: "_ga4ct_sid",
	CID_PREFIX: "GA4CT.CID.",
	SID_PREFIX: "GA4CT.SID.",

	// Helper functions
	readCookie: function (name) {
		var value = "; " + document.cookie;
		var parts = value.split("; " + name + "=");
		if (parts.length == 2) return parts.pop().split(";").shift();
	},
	writeCookie: function (name, value, expiration) {
		var d = new Date();
		d.setTime(d.getTime() + expiration);
		var expires = "expires="+ d.toUTCString();
		document.cookie = name + "=" + value + ";" + expiration + ";domain=" + ga4ct.DOMAIN + ";path=/";
	},
	// I could use UUIDs for ids which is probably fine to start
	// It might be better in the future to request IDs from the server instead?
	generateUniqueId: function() {
		return Math.floor(Math.random() * 0x7FFFFFFF) + "." + Math.floor(Date.now() / 1000);
	},

	// Main functions for managing sessions and events
	init: function () {
		
	},
	// I'll only allow event name and value for now and may extend to optional parameters later
	send: function (event, value = null) {
		
	},

	// Functions for managing device, session ID
	checkClientId: function () {
		return ga4ct.getClientId() !== undefined;
	},
	checkSessionId: function () {
		return ga4ct.getSessionId() !== undefined;
	},
	validateClientId: function () {
		return /^GA4CT\.CID\.\d+\.\d+$/.test(ga4ct.getClientId());
	},
	validateSessionId: function () {
		return /^GA4CT\.SID\.\d+\.\d+$/.test(ga4ct.getSessionId());
	},
	getClientId: function () {
		return ga4ct.readCookie(ga4ct.CID_COOKIE_NAME);
	},
	getSessionId: function () {
		return ga4ct.readCookie(ga4ct.SID_COOKIE_NAME);
	},
	newClientId: function () {
		writeCookie(ga4ct.CID_COOKIE_NAME,
			ga4ct.CID_PREFIX + ga4ct.generateUniqueId(),
			ga4ct.MAX_EPIRATION
		);
	},
	newSessionId: function () {
		writeCookie(ga4ct.CID_COOKIE_NAME,
			ga4ct.CID_PREFIX + ga4ct.generateUniqueId(),
			ga4ct.SESSION_EXPIRATION
		);
	}
}