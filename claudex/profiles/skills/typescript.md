# TypeScript Skill

<skill_expertise>
You are an expert in TypeScript, React, and modern Node.js patterns.
- **Strict Typing**: Always use strict type definitions. Avoid `any`.
- **Modern Patterns**: Use functional components, hooks, and async/await.
- **Testing**: Prioritize unit tests with Jest/Vitest.
</skill_expertise>

<coding_standards>
- Use `interface` for object definitions.
- Use `const` for immutability.
- Prefer composition over inheritance.
</coding_standards>

<best_practices>

## Type Safety
- Use strict TypeScript configuration
- Avoid `any` type - use `unknown` or proper types
- Define explicit return types for functions
- Use type guards and type predicates
- Leverage discriminated unions for state management

## Code Organization
- Use barrel exports for module organization
- Implement proper separation of concerns
- Create reusable generic types
- Use interface segregation principle
- Apply dependency injection patterns

## Testing Patterns
- Write comprehensive unit tests
- Use proper TypeScript test utilities
- Mock dependencies with type safety
- Test edge cases and error scenarios
- Maintain high test coverage

## Error Handling
- Use custom error classes
- Implement proper error boundaries
- Type error responses properly
- Use Result/Either patterns when appropriate
- Handle async errors correctly

</best_practices>

<utils>

## Test Execution Commands
Run tests with emulators and mocked services:
```bash
cd {project_path} && \
env FIRESTORE_EMULATOR_HOST=localhost:8080 \
    FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 \
    MOCK_OPENAI=true \
    NODE_OPTIONS='--experimental-vm-modules' \
    yarn jest --testPathPattern=<file_path> --testNamePattern=<name_pattern>
```

## Quality Check Commands
- `yarn fix:format` - Auto-fix code formatting issues (Prettier)
- `yarn check:types` - Validate TypeScript type safety
- `yarn check:lint` - Check code quality and style rules (ESLint)

## TypeScript-Specific Commands
- `tsc --noEmit` - Type check without emitting files
- `tsc --listFiles` - List all files included in compilation
- `tsc --showConfig` - Show resolved TypeScript configuration

</utils>

<mcp_tools>
- `mcp__context7__resolve-library-id` - Resolve library identifiers
- `mcp__context7__get-library-docs` - Get up-to-date library documentation
- `mcp__sequential-thinking__sequentialthinking` - Deep analysis for complex decisions
</mcp_tools>
