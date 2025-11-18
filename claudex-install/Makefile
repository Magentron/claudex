# BMad Installation
install:
ifndef PROJECT_PATH
	@echo "Error: PROJECT_PATH is required"
	@echo ""
	@echo "Usage: make install PROJECT_PATH=/path/to/project"
	@echo ""
	@echo "Example:"
	@echo "  make install PROJECT_PATH=/Users/username/my-project"
	@echo ""
	@exit 1
endif
	@./install.sh "$(PROJECT_PATH)"

uninstall:
ifndef PROJECT_PATH
	@echo "Error: PROJECT_PATH is required"
	@echo ""
	@echo "Usage: make uninstall PROJECT_PATH=/path/to/project"
	@echo ""
	@echo "Example:"
	@echo "  make uninstall PROJECT_PATH=/Users/username/my-project"
	@echo ""
	@exit 1
endif
	@./uninstall.sh "$(PROJECT_PATH)"

# BMad Profile Selector
bmad:
	@./bmad.sh

# BMad Agents (Direct Access)
team-lead:
	claude --system-prompt "$$(cat .claude/commands/BMad/agents/team-lead.md)" init