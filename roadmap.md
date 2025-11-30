In Progress:

To Do: 
  - Adjust Agents' models:
    - Documenter hook: haiku
    - Architect: opus
    - Researcher: haiku/sonnet
    - Engineer: sonnet/opus
    - Prompt Engineer: sonnet/opus
  - Find a solution to allow the user to define where the relevant documentation is located (product, architecture, standards, etc)
  - Review all agents to adjust their output format to make sure they provide enough context to the caller (team lead) but avoid verbose responses
  - Create QA that is responsible for defining the cases to be covered by the test and evals suite. They will receive the definition of the feature as input and come up with the test and evals suite definition at a product/business level. The QA should execute in parallel of the Architect.
  - Architect to define isolation testing strategy in execution plan:
    - Goal: Define a way to test new development in isolation to enable a quick feedback loop for the Engineer to check results and iterate.
  - Implement `claudex --update`:
    - When executed from a path where claudex was already installed, it updates configuration files (.claude, .cursor, etc).


Done:
  - **Remove Profile Selection UI** (2025-11-30):
    - Removed profile selection menus from new session and old session flows
    - Application now directly uses `team-lead-new` profile without prompting
    - Simplified user experience by eliminating unnecessary selection step
    - Net reduction of ~93 lines of UI code in main.go
  - **Composable Engineer Agents** (2025-11-29):
    - Refactored monolithic engineer agents into composable Role + Skills architecture
    - Created base `profiles/roles/engineer.md` with generic workflow and orchestration interface
    - Created language-specific skills: `typescript.md`, `python.md`, `go.md`
    - Installer dynamically detects project stack (scans subdirectories up to depth 3)
    - Installer assembles `principal-engineer-{stack}` agents at install time
    - Multi-stack projects get multiple engineers (e.g., typescript + python)
    - Empty projects prompt user to select stack(s)
    - Team-lead updated to use `principal-engineer-{stack}` with runtime stack resolution
    - Go app updated with filesystem profile discovery for dynamic agents
    - Backward compatible: existing `principal-typescript-engineer` preserved
  - Enable macOS notifications for Claude Code subagents and Claude notifications. Add voice + visual alerts when subagent tasks complete using the SubagentStop hook and when Claude sends a Notification via notification hook.
  - Refactor resume session feature:
    - Resume should: 1. ask the user if they want to start with fresh memory or continue with the previous one; 2.1 if continue is chosen the execution is like the one we have currently, 2.2 if fresh memory is chosen then a new session-id is generated, a new folder is created as a clone of the previous one and  the previous one is removed (this is what makes it different from "fork", which keeps both folders). In case of "fork", we need to refactor to allow the user to enter a description like with new sessions, with the difference that the current session folder is cloned and kept.
  - Send notification on Claude's notification hook event
