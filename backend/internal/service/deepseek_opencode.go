package service

// openAIAccountPlatformMatches permits explicitly bound OpenCode Go accounts
// to serve DeepSeek groups. Keep the account's actual platform: its protocol,
// session headers, quota windows and upstream URL must remain OpenCode's.
// The normal group binding, model permissions and scheduling gates still apply.
func openAIAccountPlatformMatches(account *Account, requestPlatform string) bool {
	if account == nil {
		return false
	}
	platform := NormalizeOpenAICompatiblePlatform(requestPlatform)
	return account.Platform == platform || (platform == PlatformDeepseek && account.Platform == PlatformOpenCodeGo)
}
