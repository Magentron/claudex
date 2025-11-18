NOTE: wondering if it would be a good idea executing `claude` in the session directory, allowing access to the repository. Why? Because it will have direct access to all the context accumulated and ignore the rest. If this context is rich it can avoid overloading the context with anything else that is in the rest of the project. Subagents would also have access to all the session's context.

- Features:
  - [ ] Create command /reload-context that: this command + the custom session management of claudex is already functional as I can start claudex with a given profile (e.g. "none" that loads claude as it is) and then use the session management with the /exit handling, resume, etc, and /reload-context.
    - Clears the current context: probably a hook that triggers before the command and deletes the "Transcript Path" (which contains the entire conversation) or maybe it can use the already built-in /clear somehow. Then, executes the command.md that instructs claude to read the context files from the session folder. Then, it loads the profile selected for that session by: post command hook that returns the profile prompt with the "active".
  - [ ] Capture when the user runs /exit:
    - Hook will summarize and create file to optimally resume the session
    - NOTE: It can work in the background! so we can do heavy stuff here without impacting the user
  - [ ] Capture when claude runs any agent (SubagentStop) to update the session context. The session context can be updated from message.stop_reason == end_turn as it contains the final result from the executed agent.
    - Not sure yet if we should capture all end_prompts submitted to update the documentation rather than only agents execution
  - [ ] Reload session's state: for instance, if we are adding additional documentation to the folder and we want Claude to load it. This should be managed via custom /command.
  - [x] Select session and agent when initiating 
  - [ ] Generic agents that can be extended with custom additional documentation
  - [ ] Enable MCPs when installing the framework
  - [ ] Run with another Coding Agent tool (e.g. codex, gemini, etc): allow the user to execute a prompt with a different CLI toolop
  - [ ] Implement hooks: every hook receives the session_id as input params, so we can inject anything we want to the prompt or do anything we want. Examples:
    - [ ] preCompact: when the compact is going to execute, intercept and update the session data, then abort. Restart the agent to load the session data from scratch
    - [ ] for every command, pass the session path so the agent always knows that it has to refer to this context. For instance, useful when spawning new agents so they have all relevant context. Prompt: "you are working on this workspace: /path/to/session"
    - [ ] I can add custom commands (i.e. /gemini, /codex, etc) so that when the user executes them, the script hook builds a great prompt and executes these other tools to have a second opinion on something. The result can be passed again to claude.

- Installation script should:
  - create symbolik links from .claudex to the corresponding .claude folders
  - create symbolik links from .claudex to the corresponding .cursor folders
  - create symbolik links from project path to .claude, .cursor and additional required folders
  - copy and replace the claudex script to $PATH

- .claudex structure:
  - agents
  - tasks
  - templates
  - sessions: every time the user runs `claudex`, they will be prompted to create a new session or resume an existing one. The user should be able to navigate with the arrow keys, being the first option the creation where the user can enter a description of the session which will be used to generate the name of the session and create a new folder under sessions/ with that name (claude should be used for this purpose). Second option should be `ephemeral` which won't create a new folder and does not require description.

- Think about the way that the user will be able to inject documentation:

IMPORTANT:
session_id = claude --system-prompt "prompt" "activate" --output-format json | jq -r '.session_id'
claude --resume $session_id




FEEDBACK:
- Currently, Claude can generate documents related to the ongoing work anywhere. It should stick to the session's folder.
- The current Team Lead seems to do a great job at delegating the work to the Architect Assistant and Engineer ✅