# Clarification Workflow

## Purpose

This document provides a quick reference for the mandatory clarification phase that must precede all planning work (execution plans, architecture documents, refactoring plans).

## The Two-Phase Approach

### Phase 1: Mandatory Clarification Phase

**MUST be completed before ANY document creation**

### Phase 2: Document Creation

**ONLY starts after Phase 1 is complete**

## Phase 1: Step-by-Step Process

### Step 1: Acknowledge and Set Context

- Acknowledge the user's request
- Briefly summarize what you understand they want to accomplish
- State explicitly: "Before I create the execution plan, I need to clarify some details to ensure we're aligned."

### Step 2: Gather Context

- Read any provided PRD, architecture documents, user stories, or requirements
- Search the codebase for relevant existing implementations
- Identify all libraries, SDKs, frameworks, third-party services mentioned

### Step 3: Query Documentation (MANDATORY)

- Use `mcp__context7__resolve-library-id` to identify libraries
- Use `mcp__context7__get-library-docs` to fetch up-to-date documentation
- Examples: Firebase Functions, OpenAI SDK, TypeScript, React, etc.

### Step 4: Analyze Complexity (when applicable)

- Use `mcp__sequential-thinking__sequentialthinking` for:
  - Complex architectural decisions
  - Multiple alternative approaches
  - Interconnected system changes
  - Trade-off analysis

### Step 5: Ask Clarifying Questions (CRITICAL)

Present ALL clarifying questions in a single, organized message:

**Format:**

```markdown
Before I create the execution plan, I need to clarify the following:

## Architecture Questions

1. [Question about architectural approach]
2. [Question about system design]

## Implementation Details

3. [Question about specific implementation]
4. [Question about technology choice]

## Testing Strategy

5. [Question about test coverage]
6. [Question about test approach]

## [Other Categories as needed]

...
```

**Topics to Cover:**

- ANY unclear technical decisions
- Implementation approaches that need confirmation
- Assumptions that need verification
- Ambiguities in requirements
- Technology choices or patterns to follow
- Performance requirements
- Security considerations
- Error handling strategies
- Testing expectations
- Timeline or phasing considerations

**WAIT for user responses before proceeding**

### Step 6: Review User Responses

- If any answers create new questions, ask follow-up questions
- If any answers are unclear, ask for clarification
- Repeat until ALL questions are resolved
- **DO NOT proceed until user explicitly confirms everything is clear**

### Step 7: Confirm Understanding

Present a summary:

```markdown
Based on your answers, here's my understanding of what we'll implement:

## Summary of Decisions

### [Category 1]

- [Decision 1]
- [Decision 2]

### [Category 2]

- [Decision 3]
- [Decision 4]

### [Continue for all categories]

Does this summary accurately reflect what we want to implement? Should I proceed with creating the execution plan?
```

**WAIT for explicit user confirmation**

### Step 8: Lock in Final Decisions

- Once user confirms, you now have all the information needed
- You are now ready to create the document with ONLY final decisions
- Do NOT include alternative options in the document
- Do NOT include rationale discussions in the document
- Keep document focused on what will be implemented

## Success Criteria

### Phase 1 is complete when:

✅ All clarifying questions have been asked
✅ User has answered all questions
✅ Follow-up questions have been resolved
✅ Final decisions have been summarized
✅ User has explicitly confirmed the summary
✅ User has approved proceeding to document creation

### DO NOT proceed to Phase 2 if:

❌ Any questions remain unanswered
❌ Any requirements are still ambiguous
❌ User has not confirmed the summary
❌ User has not explicitly approved proceeding

## Common Mistakes to Avoid

1. **Skipping the clarification phase** - Always start with clarification, even if requirements seem clear
2. **Starting document creation too early** - Wait for explicit user approval
3. **Asking questions while creating documents** - All questions should be in Phase 1
4. **Including alternatives in documents** - Documents should only contain final decisions
5. **Not using context7 MCP** - Always query up-to-date documentation
6. **Not summarizing decisions** - Always provide a summary for user confirmation

## Example Interaction Flow

**User:** "I need to add user preference caching to our Firebase Functions"

**Architect (Phase 1 - Step 1):**
"I understand you want to add user preference caching to the Firebase Functions. Before I create the execution plan, I need to clarify some details to ensure we're aligned."

**Architect (Phase 1 - Steps 2-4):**
[Gathers context, queries documentation, analyzes complexity]

**Architect (Phase 1 - Step 5):**
"Before I create the execution plan, I need to clarify the following:

## Architecture Questions

1. What caching strategy do you prefer: in-memory, Redis, or database-backed?
2. Should the cache be per-function-instance or shared across instances?
   [... more questions ...]"

**User:** [Provides answers]

**Architect (Phase 1 - Step 6):**
[May ask follow-up questions if needed]

**Architect (Phase 1 - Step 7):**
"Based on your answers, here's my understanding:

## Summary of Decisions

[... comprehensive summary ...]

Does this summary accurately reflect what we want to implement? Should I proceed with creating the execution plan?"

**User:** "Yes, that's correct. Please proceed."

**Architect (Phase 1 - Step 8):**
"Perfect! I'll now create the execution plan based on these decisions."

**Architect (Phase 2):**
[Begins document creation with final decisions only]

## Integration with Other Tasks

This clarification workflow applies to:

- Execution plan creation (`create-execution-plan.md`)
- Architecture document creation
- Refactoring plans
- Any planning work that results in a formal document

The workflow does NOT apply to:

- Code implementation (that's for the Dev agent)
- Quick clarification conversations
- Status updates or questions
