<!-- Powered by BMAD™ Core -->

# Create Execution Plan Document

## ⚠️ CRITICAL EXECUTION NOTICE ⚠️

**THIS IS AN EXECUTABLE WORKFLOW - NOT REFERENCE MATERIAL**

When this task is invoked:

1. **MANDATORY CLARIFICATION PHASE FIRST** - You MUST start with an explicit clarification phase before ANY document creation
2. **CLARIFY BEFORE DOCUMENTING** - Resolve ALL questions with the user before producing document sections
3. **USE MCP TOOLS** - Query documentation via context7 MCP, use sequential-thinking for complex analysis
4. **FINAL DECISIONS ONLY** - Documents must contain ONLY final decisions, not alternatives or rationale discussions
5. **MANDATORY STEP-BY-STEP EXECUTION** - Each section must be processed sequentially with user feedback
6. **ELICITATION IS REQUIRED** - When `elicit: true`, you MUST use the 1-9 format and wait for user response

**VIOLATION INDICATOR:** If you create document sections with alternatives, options, or rationale before clarifying with user, you have violated this workflow.

**WORKFLOW VIOLATION:** Starting document creation before completing the explicit clarification phase violates this task.

## PHASE 1: MANDATORY CLARIFICATION PHASE

**THIS PHASE MUST BE COMPLETED BEFORE ANY DOCUMENT CREATION BEGINS**

### Overview

When the user requests an execution plan (for a new feature, refactoring, or any code work), you MUST start with this explicit clarification phase. The goal is to ensure complete understanding and alignment BEFORE creating any document sections.

### Step-by-Step Clarification Process

**Step 1: Acknowledge and Set Context**

- Acknowledge the user's request
- Briefly summarize what you understand they want to accomplish
- State explicitly: "Before I create the execution plan, I need to clarify some details to ensure we're aligned."

**Step 2: Gather Context**

- Read any provided PRD, architecture documents, user stories, or requirements
- Search the codebase for relevant existing implementations
- Identify all libraries, SDKs, frameworks, third-party services mentioned

**Step 3: Query Documentation** (MANDATORY)

- Use `mcp__context7__resolve-library-id` to identify libraries
- Use `mcp__context7__get-library-docs` to fetch up-to-date documentation
- Examples: Firebase Functions, OpenAI SDK, TypeScript, React, etc.

**Step 4: Analyze Complexity** (when applicable)

- Use `mcp__sequential-thinking__sequentialthinking` for:
  - Complex architectural decisions
  - Multiple alternative approaches
  - Interconnected system changes
  - Trade-off analysis

**Step 5: Ask Clarifying Questions** (CRITICAL)

**MANDATORY: Use AskUserQuestion tool**

- **ALWAYS present clarifying questions using AskUserQuestion tool**
- Break down complex requirements into logical categories as **interactive menus with numbered options** or **tabbed interfaces**
- **WAIT for user responses before proceeding**

**Step 6: Review User Responses**

- If any answers create new questions, ask follow-up questions
- If any answers are unclear, ask for clarification
- Repeat until ALL questions are resolved
- **DO NOT proceed until user explicitly confirms everything is clear**

**Step 7: Confirm Understanding**

- Summarize the final decisions based on user's answers
- Present this summary to the user for final confirmation
- Ask: "Does this summary accurately reflect what we want to implement? Should I proceed with creating the execution plan?"
- **WAIT for explicit user confirmation**

**Step 8: Lock in Final Decisions**

- Once user confirms, you now have all the information needed
- Document ONLY the final decisions in the execution plan
- Do NOT include alternative options in the document
- Do NOT include rationale discussions in the document
- Keep document focused on what will be implemented

### Clarification Phase Success Criteria

✅ The clarification phase is complete when:

- All technical questions have been presented using interactive UI components
- User has engaged with the interactive menus and provided answers
- All ambiguities have been resolved
- All assumptions have been verified
- User has confirmed the summary of decisions
- User has given explicit permission to proceed with execution plan creation

❌ DO NOT proceed to document creation if:

- Questions were presented as plain text instead of interactive UI
- Any questions remain unanswered
- Any requirements are still ambiguous
- User has not confirmed the summary
- User has not explicitly approved proceeding

---

## PHASE 2: DOCUMENT CREATION

**ONLY START THIS PHASE AFTER COMPLETING PHASE 1**

## MCP Tool Usage Examples

### Using context7 for Documentation

```
# Step 1: Resolve library ID
Use: mcp__context7__resolve-library-id
Input: "firebase-functions"
Output: Library ID for Firebase Functions

# Step 2: Get documentation
Use: mcp__context7__get-library-docs
Input: {library_id: "...", query: "how to create http callable functions"}
Output: Up-to-date documentation about callable functions
```

### Using sequential-thinking for Complex Analysis

```
Use: mcp__sequential-thinking__sequentialthinking
Input: {
  "task": "Analyze the best approach for implementing user preference caching with multiple storage options (Redis, in-memory, database)",
  "context": "Firebase Cloud Functions with cold start concerns, budget constraints, need for consistency"
}
Output: Structured thinking process with step-by-step analysis
```

## CRITICAL: Mandatory Elicitation Format

**When `elicit: true`, this is a HARD STOP requiring user interaction:**

**YOU MUST:**

1. Present section content
2. Provide detailed rationale (explain trade-offs, assumptions, decisions made)
3. **STOP and present numbered options 1-9:**
   - **Option 1:** Always "Proceed to next section"
   - **Options 2-9:** Select 8 methods from data/elicitation-methods
   - End with: "Select 1-9 or just type your question/feedback:"
4. **WAIT FOR USER RESPONSE** - Do not proceed until user selects option or provides feedback

**WORKFLOW VIOLATION:** Creating content for elicit=true sections without user interaction violates this task.

**NEVER ask yes/no questions or use any other format.**

## Processing Flow

### PHASE 1: Mandatory Clarification Phase

1. **Acknowledge Request** - Summarize understanding and state need for clarification
2. **Gather Context** - Read all relevant documents and codebase
3. **Query Documentation** - Use context7 MCP for library/framework docs
4. **Analyze Complexity** - Use sequential-thinking MCP for complex decisions
5. **Ask Clarifying Questions** - Present all questions organized by category
6. **Review Responses** - Ask follow-ups until everything is clear
7. **Confirm Understanding** - Summarize decisions and get explicit approval
8. **Lock in Decisions** - Ready to create document with final decisions only

### PHASE 2: Document Creation (Only After Phase 1 Complete)

1. **Load Template** - Use execution-plan-tmpl.yaml
2. **Set Preferences** - Show current mode (Interactive), confirm output file
3. **Process Each Section:**
   - Skip if condition unmet
   - Use context7 MCP when referencing library/framework documentation
   - Use sequential-thinking MCP for complex implementation decisions
   - Draft content using section instruction (using ONLY final decisions from Phase 1)
   - Present content + detailed rationale
   - **IF elicit: true** → MANDATORY 1-9 options format
   - Save to file if possible
4. **Continue Until Complete**

## Detailed Rationale Requirements

When presenting section content, ALWAYS include rationale that explains:

- Trade-offs and choices made (what was chosen over alternatives and why)
- Key assumptions made during drafting
- Interesting or questionable decisions that need user attention
- Areas that might need validation

## Elicitation Results Flow

After user selects elicitation method (2-9):

1. Execute method from data/elicitation-methods
2. Present results with insights
3. Offer options:
   - **1. Apply changes and update section**
   - **2. Return to elicitation menu**
   - **3. Ask any questions or engage further with this elicitation**

## Test Execution Command Format

**CRITICAL:** Always use this exact format for test commands:

```bash
cd /Users/maikel/Workspace/Pelago/voiced/pelago/apps/voiced/functions && env FIRESTORE_EMULATOR_HOST=localhost:8080 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 MOCK_OPENAI=true NODE_OPTIONS='--experimental-vm-modules' yarn jest --testPathPattern=<file_path> --testNamePattern=<name_pattern>
```

Where:

- `<file_path>` contains the test file path to be executed
- `<name_pattern>` allows you to execute a subset of tests

## YOLO Mode

User can type `#yolo` to toggle to YOLO mode (process all sections at once).

## CRITICAL REMINDERS

**❌ NEVER:**

- **Start document creation without completing Phase 1 clarification**
- **Skip the explicit clarification phase**
- **Present clarification questions as a wall of text - always use interactive UI components**
- Create document sections before clarifying all questions with user
- Proceed without user's explicit approval after summarizing decisions
- Include alternatives, options, or rationale discussions in final document
- Skip using context7 MCP when documentation queries are needed
- Ask yes/no questions for elicitation
- Use any format other than 1-9 numbered options
- Create new elicitation methods

**✅ ALWAYS:**

- **Begin with Phase 1: Mandatory Clarification Phase**
- **Present all clarifying questions using interactive UI components (numbered menus, tabbed interfaces)**
- **Break down complex requirements into logical categories presented step-by-step**
- **Get explicit user confirmation before starting document creation**
- Use context7 MCP to query up-to-date documentation
- Use sequential-thinking MCP for complex analysis
- Clarify ALL questions before producing document sections
- Summarize final decisions and get user approval
- Document ONLY final decisions in the execution plan
- Use exact 1-9 format when elicit: true
- Select options 2-9 from data/elicitation-methods only
- Provide detailed rationale explaining decisions
- End with "Select 1-9 or just type your question/feedback:"
- Use exact test command format as specified

## Success Criteria

### Phase 1 (Clarification) is complete when:

1. All clarifying questions have been presented using interactive UI components
2. User has answered all questions through the interactive interface
3. Follow-up questions have been resolved
4. Final decisions have been summarized
5. User has explicitly confirmed the summary
6. User has approved proceeding to document creation

### Phase 2 (Document) is complete when:

1. Phase 1 has been successfully completed
2. Document contains ONLY final decisions (no alternatives or rationale)
3. Executive Summary clearly states what, why, and how
4. Implementation Overview includes high-level flow and code changes
5. Test suite is fully defined with exact test commands
6. File-by-file implementation provides clear guidance
7. Code quality checks are specified
8. Implementation checklist breaks work into actionable tasks
9. All relevant documentation was queried via context7 MCP
10. Complex decisions were analyzed via sequential-thinking MCP
