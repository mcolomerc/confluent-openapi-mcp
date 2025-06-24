#!/bin/bash

# Script to restart VS Code MCP servers
echo "🔧 Build completed! MCP server binary updated."
echo ""
echo "📋 To restart VS Code MCP servers:"
echo "   1. Press Cmd+Shift+P to open Command Palette"
echo "   2. Type 'MCP: Restart All Servers'"
echo "   3. Press Enter"
echo ""
echo "🔄 Alternatively, reload VS Code window:"
echo "   1. Press Cmd+Shift+P"
echo "   2. Type 'Developer: Reload Window'"
echo "   3. Press Enter"
echo ""

# Try to bring VS Code to front
osascript -e 'tell application "Visual Studio Code" to activate' 2>/dev/null || true

echo "✅ VS Code should now be in focus. Ready to restart MCP servers!"
