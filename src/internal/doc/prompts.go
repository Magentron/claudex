package doc

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
)

// LoadPromptTemplate reads a prompt template file from the filesystem
func LoadPromptTemplate(fs afero.Fs, templatePath string) (string, error) {
	data, err := afero.ReadFile(fs, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("prompt template is empty: %s", templatePath)
	}

	return content, nil
}

// BuildDocumentationPrompt combines template with transcript content and session context
// The template can contain placeholders that will be replaced with actual values:
// - $RELEVANT_CONTENT: The formatted transcript content
// - $DOC_CONTEXT: Session context and existing documentation
// - $SESSION_FOLDER: Absolute path to the session folder
func BuildDocumentationPrompt(template string, transcriptContent string, sessionContext string, sessionFolder string) string {
	prompt := template

	// Replace placeholders with actual content
	prompt = strings.ReplaceAll(prompt, "$RELEVANT_CONTENT", transcriptContent)
	prompt = strings.ReplaceAll(prompt, "$DOC_CONTEXT", sessionContext)
	prompt = strings.ReplaceAll(prompt, "$SESSION_FOLDER", sessionFolder)

	return prompt
}
