# Directives

This folder contains directive files that define system-level instructions and guardrails for the MCP server.

## Files

- **`role.txt`** - Defines the assistant's role and expertise area
- **`guardrails.txt`** - Security and behavioral constraints  
- **`validation.txt`** - Operation validation requirements

## Usage

These files are automatically:

1. Copied to `bin/directives/` during the build process
2. Loaded by the MCP server at startup
3. Prepended to all prompt responses

## Format

Each file should contain plain text instructions, one per line or in paragraph form. Multiple files are combined with double newlines between them.

## Editing

To modify directives:

1. Edit the appropriate `.txt` file in this folder
2. Run `make build` to copy changes to the binary directory
3. Restart the MCP server to load the new directives
