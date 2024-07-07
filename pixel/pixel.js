window.ga4ct = {
	// Helpful values
	SESSION_EXPIRATION: 1800000, // 30 minutes in milliseconds
	MAX_EPIRATION: 46656000000, // 540 days in milliseconds
	DOMAIN: document.location.hostname.split('.').reverse().splice(0,2).reverse().join('.'), // root domain
	CID_COOKIE_NAME: "_ga4ct_cid",
	SID_COOKIE_NAME: "_ga4ct_sid",
	CID_PREFIX: "GA4CT.CID.",
	SID_PREFIX: "GA4CT.SID.",
	CID_REGEX: /^GA4CT\.CID\.\d+\.\d+$/,
	SID_REGEX: /^GA4CT\.SID\.\d+\.\d+$/,
	AID_REGEX: /^GA4CT-\d+$/,
	SEND_ENDPOINT: "http://localhost:3000/event",

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
		document.cookie = name + "=" + value + ";" + expires + ";domain=" + ga4ct.DOMAIN + ";path=/";
	},
	// I could use UUIDs for ids which is probably fine to start
	// It might be better in the future to request IDs from the server instead?
	generateUniqueId: function() {
		return Math.floor(Math.random() * 0x7FFFFFFF) + "." + Math.floor(Date.now() / 1000);
	},

	init: function (account_id) {
		ga4ct.ACCOUNT_ID = account_id;
		if(!ga4ct.validateAccountId()) {
			console.log("GA4CT Error: Invalid Account ID format.")
		}
	},
	// I'll only allow event name and value for now and may extend to optional parameters later
	send: function (event_name, event_value = 0) {
		if(!ga4ct.validateAccountId()) {
			console.log("GA4CT Error: Initialize GA4CT before sending events.");
			return
		}

		// Check if CID and SID are set. If not, set new ones
		if(!ga4ct.validateClientId()) {
			ga4ct.newClientId();
			ga4ct.newSessionId();
		} else if (!ga4ct.validateSessionId()) {
			ga4ct.extendCliendId();
			ga4ct.newSessionId();
		} else {
			ga4ct.extendCliendId();
			ga4ct.extendSessionId();
		}

		// If cookies were unable to be written, don't send events.
		if(!ga4ct.validateClientId() || !ga4ct.validateSessionId()) {
			console.log("GA4CT Error: Client ID or Session ID is not set.");
			return
		}

		fetch(ga4ct.SEND_ENDPOINT, {
			method: "POST",
			headers: {
				"Content-Type": "application/json"
			},
			body: JSON.stringify({
				"account_id": ga4ct.ACCOUNT_ID,
				"client_id": ga4ct.getClientId(),
				"session_id": ga4ct.getSessionId(),
				"event_name": event_name, 
				"event_value": event_value, 
				"timestamp": new Date().toISOString(),
				"page_location": document.location.href,
				"page_title": document.title,
				"page_referrer": document.referrer,
				"user_agent": navigator.userAgent,
				"screen_resolution": window.screen.width + "x" + window.screen.height
			})
		})
	},

	// Functions for managing device, session ID
	validateAccountId: function () {
		return ga4ct.AID_REGEX.test(ga4ct.ACCOUNT_ID);
	},
	validateClientId: function () {
		return ga4ct.CID_REGEX.test(ga4ct.getClientId());
	},
	validateSessionId: function () {
		return ga4ct.SID_REGEX.test(ga4ct.getSessionId());
	},
	getClientId: function () {
		return ga4ct.readCookie(ga4ct.CID_COOKIE_NAME);
	},
	getSessionId: function () {
		return ga4ct.readCookie(ga4ct.SID_COOKIE_NAME);
	},
	newClientId: function () {
		ga4ct.writeCookie(ga4ct.CID_COOKIE_NAME,
			ga4ct.CID_PREFIX + ga4ct.generateUniqueId(),
			ga4ct.MAX_EPIRATION
		);
	},
	newSessionId: function () {
		ga4ct.writeCookie(ga4ct.SID_COOKIE_NAME,
			ga4ct.SID_PREFIX + ga4ct.generateUniqueId(),
			ga4ct.SESSION_EXPIRATION
		);
	},
	extendCliendId: function() {
		ga4ct.writeCookie(ga4ct.CID_COOKIE_NAME,
			ga4ct.getClientId(),
			ga4ct.MAX_EPIRATION
		);
	},
	extendSessionId: function () {
		ga4ct.writeCookie(ga4ct.SID_COOKIE_NAME,
			ga4ct.getSessionId(),
			ga4ct.SESSION_EXPIRATION
		);
	}
}