user_pref("browser.aboutwelcome.enabled", false);
user_pref("browser.startup.homepage", "about:home");

// AppAutoUpdate: false
user_pref("app.update.auto", false);

// AutofillAddressEnabled: false
user_pref("extensions.formautofill.addresses.enabled", false);

// AutofillCreditCardEnabled: false
user_pref("extensions.formautofill.creditCards.enabled", false);

// NoDefaultBookmarks: true
user_pref("browser.bookmarks.restore_default_bookmarks", false);

// DisableFirefoxAccounts: true
user_pref("identity.fxaccounts.enabled", false);

// DisableFirefoxStudies: true
user_pref("app.shield.optoutstudies.enabled", false);
user_pref("app.normandy.enabled", false);

// DisableFormHistory: true
user_pref("browser.formfill.enable", false);

// DisablePocket: true
user_pref("extensions.pocket.enabled", false);

// DisableTelemetry: true
user_pref("datareporting.policy.dataSubmissionEnabled", false);
user_pref("datareporting.healthreport.uploadEnabled", false);
user_pref("toolkit.telemetry.enabled", false);
user_pref("toolkit.telemetry.unified", false);

// DisplayBookmarksToolbar: true
user_pref("browser.toolbars.bookmarks.visibility", "always");

// DontCheckDefaultBrowser: true
user_pref("browser.shell.checkDefaultBrowser", false);

// HomePage: "none" and StartPage: "none"
user_pref("browser.startup.homepage", "about:blank");
user_pref("browser.startup.page", 0);

// NewTabPage: true (Standard blank/active newtab)
user_pref("browser.newtabpage.enabled", true);

// OfferToSaveLogins / PasswordManagerEnabled: false
user_pref("signon.rememberSignons", false);

// OverrideFirstRunPage / OverridePostUpdatePage
user_pref("startup.homepage_welcome_url", "");
user_pref("startup.homepage_welcome_url.additional", "");
user_pref("browser.startup.homepage_override.mstone", "ignore");

// PrivateBrowsingModeAvailability: 2 (Force Private Browsing)
user_pref("browser.privatebrowsing.autostart", true);

// SanitizeOnShutdown: true
user_pref("privacy.sanitize.sanitizeOnShutdown", true);
user_pref("privacy.clearOnShutdown.cache", true);
user_pref("privacy.clearOnShutdown.cookies", true);
user_pref("privacy.clearOnShutdown.downloads", true);
user_pref("privacy.clearOnShutdown.formdata", true);
user_pref("privacy.clearOnShutdown.history", true);
user_pref("privacy.clearOnShutdown.offlineApps", true);
user_pref("privacy.clearOnShutdown.sessions", true);

// SkipTermsOfUse: true
user_pref("browser.rights.maybeAtFirstRun", false);

// UserMessaging Settings
user_pref("browser.aboutwelcome.enabled", false);
user_pref("browser.messaging-system.whatsNewPanel.enabled", false);
user_pref("browser.vpn_promo.enabled", false);
user_pref("extensions.htmlaboutaddons.recommendations.enabled", false);
user_pref("browser.discovery.enabled", false);
