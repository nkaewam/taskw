# Documentation TODO List

This document outlines all the documentation pages needed for the Taskw project using Fumadocs.

## 📚 Core Documentation Structure

### 1. Getting Started

- [ ] **overview.mdx** - Project overview, what problems it solves, key features
- [ ] **installation.mdx** - Installation methods (go install, binary downloads, build from source)
- [ ] **quick-start.mdx** - 5-minute tutorial from zero to working API
- [ ] **migration.mdx** - How to migrate existing Fiber/Wire projects to use Taskw

### 2. CLI Reference

- [ ] **cli/index.mdx** - CLI overview and common patterns
- [ ] **cli/init.mdx** - `taskw init` command documentation
- [ ] **cli/generate.mdx** - `taskw generate` and its subcommands (all, routes, deps)
- [ ] **cli/scan.mdx** - `taskw scan` command for debugging and preview
- [ ] **cli/clean.mdx** - `taskw clean` command for cleanup
- [ ] **cli/flags.mdx** - Global flags and configuration options

### 3. Configuration

- [ ] **config/index.mdx** - Configuration overview and structure
- [ ] **config/taskw-yaml.mdx** - Complete taskw.yaml reference
- [ ] **config/project-setup.mdx** - Project structure recommendations
- [ ] **config/paths.mdx** - Path configuration and scanning directories
- [ ] **config/generation.mdx** - Generation options and output configuration

### 4. Core Concepts

- [ ] **concepts/index.mdx** - How Taskw works (scanning → analysis → generation)
- [ ] **concepts/annotations.mdx** - Swaggo @Router annotations and patterns
- [ ] **concepts/providers.mdx** - Provider function patterns and Wire integration
- [ ] **concepts/handlers.mdx** - Handler function structure and best practices
- [ ] **concepts/code-generation.mdx** - What gets generated and why
- [ ] **concepts/file-watching.mdx** - Development workflow and file watching

## 📋 Content Guidelines

### Writing Style

- Use clear, concise language appropriate for developers
- Include practical examples for every concept
- Start with simple examples, then progress to complex ones
- Always show both the code and the expected output
- Use consistent terminology throughout all docs

### Code Examples

- All Go code should be production-ready and follow best practices
- Include complete, runnable examples where possible
- Show both manual and Taskw-generated approaches for comparison
- Add comments explaining non-obvious parts
- Test all examples before publishing

### Documentation Structure

- Start each page with a brief overview
- Use clear headings and subheadings
- Include a "Next Steps" section pointing to related topics
- Add callouts for important warnings or tips
- End complex topics with a summary

## 📝 Notes

- **Target Audience**: Go developers familiar with Fiber and basic DI concepts
- **Tone**: Professional but friendly, focusing on practical solutions
- **Update Strategy**: Keep docs in sync with code changes, especially for CLI and config
- **Validation**: Each example should be tested against the latest version of Taskw
