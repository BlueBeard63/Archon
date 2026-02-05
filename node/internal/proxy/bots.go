package proxy

// DefaultBotUserAgents returns the default list of bot user agent patterns
// These are common search engine and social media bots that may need
// to be redirected to a pre-rendering service for SEO purposes
func DefaultBotUserAgents() []string {
	return []string{
		// Search Engine Bots
		"Googlebot",
		"Bingbot",
		"Slurp",          // Yahoo
		"DuckDuckBot",
		"Baiduspider",
		"YandexBot",
		"Sogou",
		"Exabot",
		"facebot",        // Facebook crawler
		"ia_archiver",    // Alexa

		// Social Media Bots
		"Twitterbot",
		"LinkedInBot",
		"Pinterest",
		"WhatsApp",
		"TelegramBot",
		"Slackbot",
		"Discordbot",
		"vkShare",

		// Other Common Bots
		"Applebot",
		"SemrushBot",
		"AhrefsBot",
		"MJ12bot",
		"Screaming Frog",
		"rogerbot",
		"embedly",
		"Quora Link Preview",
		"showyoubot",
		"outbrain",
		"W3C_Validator",
	}
}

// EscapeNginxRegex escapes special regex characters for nginx map directive
// The map directive uses a simple pattern matching, not full regex
func EscapeNginxRegex(pattern string) string {
	// For nginx map with ~* prefix, we just need to escape literal dots
	// Most bot names don't need escaping
	return pattern
}

// BuildNginxBotMapEntries builds the map entries for nginx bot detection
func BuildNginxBotMapEntries(userAgents []string) string {
	if len(userAgents) == 0 {
		userAgents = DefaultBotUserAgents()
	}

	result := ""
	for _, ua := range userAgents {
		// Use case-insensitive regex match with ~*
		result += "        ~*" + EscapeNginxRegex(ua) + " 1;\n"
	}
	return result
}
