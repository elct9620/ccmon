# ccmon Development Log Generator

Create a comprehensive development log for today in English, following this standardized format and guidelines. The document should be readable in 5 minutes (1200-1500 words maximum) and serve as a historical record for future development decisions.

## Output Requirements

**File Location**: `docs/devlog/[YYYYMMDD].md` where `[YYYYMMDD]` is today's date (e.g., `20250702.md`)

**Structure**: Use exactly this template structure:

```markdown
# Development Log - [Date in YYYY-MM-DD format]

## What's New
[New features, capabilities, and enhancements that provide value to users or developers]

## What's Fixed  
[Bug fixes, issues resolved, and improvements to existing functionality]

## Design Decisions
[Architectural choices, alternatives considered, and reasoning for future reference]

## Impact
[Brief summary of how these changes affect users, developers, or the project direction]

## Files Modified
[Bulleted list of key files changed for quick reference]
```

## Content Guidelines

### What's New Section
- **Focus on**: New features, enhancements, added capabilities
- **Format**: Use descriptive headers (####) for major features
- **Include**: User impact, developer benefits, new possibilities enabled
- **Exclude**: Implementation details, code snippets (unless essential for understanding)

### What's Fixed Section  
- **Structure**: Problem description → Root cause (if relevant) → Solution implemented
- **Focus on**: User-visible improvements, stability enhancements, developer experience
- **Include**: Impact of the fix, scenarios that now work correctly
- **Highlight**: Critical bugs, performance improvements, reliability enhancements

### Design Decisions Section
- **Document**: Architectural choices, pattern selections, technology decisions
- **Include**: 
  - Problem context and alternatives considered
  - Reasoning and trade-offs made
  - Future implications and guidance for similar decisions
- **Format**: Decision title → Context → Choice → Rationale
- **Critical**: Any decisions that change development patterns or conflict with existing approaches

### Quality Standards
- **Reading Time**: 5 minutes maximum (1200-1500 words)
- **Audience**: Developers familiar with the project
- **Tone**: Professional, clear, technically accurate but accessible
- **Depth**: Enough detail to understand decisions, not implementation specifics

## Research Instructions

1. **Check recent commits**: Use `git log --oneline -10` to identify changes since last devlog
2. **Analyze changes**: Look at modified files, commit messages, and change patterns
3. **Categorize changes**: Group related commits and changes by the three main sections
4. **Extract decisions**: Identify any architectural, design, or technical choices made
5. **Check conflicts**: Review if any decisions contradict current CLAUDE.md guidance

## CLAUDE.md Integration

**Critical**: If any design decisions documented in the devlog contradict or change the development approach described in CLAUDE.md:

1. **Update CLAUDE.md** to reflect the new approach
2. **Document the change** in the devlog's Design Decisions section
3. **Explain reasoning** for why the previous approach was modified
4. **Ensure consistency** between devlog and CLAUDE.md for future development

## File Analysis Guidelines

When analyzing git changes:
- **Focus on intent**: What problem was being solved?
- **Identify patterns**: Are there recurring themes in the changes?
- **Note architecture changes**: Any new patterns, refactoring, or structural improvements?
- **Capture rationale**: Why were specific approaches chosen?

## Final Checklist

Before saving the devlog, verify:
- [ ] All three main sections are present and populated
- [ ] Word count is under 1500 words (5-minute read)
- [ ] Design decisions include reasoning and context
- [ ] File list captures major changes
- [ ] CLAUDE.md is updated if needed
- [ ] Date format is correct (YYYY-MM-DD)
- [ ] Content provides value for future developers reviewing project history

This devlog will serve as a crucial historical record for understanding project evolution and making informed decisions in future development.
