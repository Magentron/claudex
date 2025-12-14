package updatecheck

type Result int

const (
	ResultNeverAskAgain Result = iota // User opted out permanently
	ResultUpToDate                    // Current version >= latest
	ResultCached                      // Cache valid, no new version
	ResultNetworkError                // Failed to check, skip silently
	ResultPromptUser                  // New version available
)
