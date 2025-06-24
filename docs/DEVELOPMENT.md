# Development Guide

This guide explains how to develop with the MCP Server, including auto-rebuild functionality, server modes, and best practices.

## 🚀 Quick Start

```bash
# 1. Install development tools
make install-tools

# 2. Start development server with auto-reload
./scripts/dev-start.sh
# OR use VS Code task: "Dev: Start Auto-Reload Server"
```

## 🔧 Server Modes

The MCP server supports three different modes to avoid port conflicts and enable efficient development:

### **HTTP Mode** (Development)
```bash
./bin/mcp-server --mode http --env .env
```
- ✅ **Purpose**: Development and testing via HTTP API
- ✅ **Port**: Runs on `:8080`
- ✅ **API**: REST endpoints at `/tools` and `/invoke`
- ✅ **Auto-reload**: Supported via Air

### **STDIO Mode** (VS Code MCP)
```bash
./bin/mcp-server --mode stdio --env .env
```
- ✅ **Purpose**: VS Code MCP integration
- ✅ **Protocol**: JSON-RPC over STDIO
- ✅ **Connection**: Persistent connection managed by VS Code
- ⚠️ **Auto-reload**: Manual restart required via VS Code

### **Both Mode** (Default)
```bash
./bin/mcp-server --mode both --env .env
```
- ✅ **Purpose**: Runs both HTTP and STDIO servers
- ⚠️ **Conflicts**: Can cause port conflicts in development
- ❌ **Recommendation**: Use specific modes instead

## 📦 Auto-Rebuild System

### **How It Works**
The auto-rebuild system uses **Air** to watch for file changes and automatically:

1. **Detects Changes**: Monitors `.go`, `.json`, `.env` files
2. **Builds HTTP Binary**: Creates `./tmp/main` for HTTP server
3. **Updates MCP Binary**: Copies to `./bin/mcp-server` for VS Code
4. **Restarts HTTP Server**: Only HTTP mode server restarts automatically
5. **Shows Instructions**: Displays how to restart VS Code MCP server

### **What Gets Auto-Rebuilt**
```
File Change → Air Detects → Rebuilds Both:
├── ./tmp/main (HTTP dev server) ← Automatically restarts
└── ./bin/mcp-server (VS Code MCP) ← Manual restart needed
```

### **Development Workflow**
1. **Start Development**: `./scripts/dev-start.sh` (HTTP-only mode)
2. **Make Code Changes**: Edit Go files, Air detects changes
3. **HTTP Server**: Automatically restarts with changes
4. **VS Code MCP**: Follow restart instructions when prompted
5. **Test Both**: HTTP API + VS Code MCP integration

## 🛠️ Development Options

### **Option 1: Smart Development Script (Recommended)**

```bash
./scripts/dev-start.sh
```

**Features:**
- ✅ **Auto-cleanup**: Kills existing servers and cleans ports
- ✅ **HTTP-only mode**: Avoids port conflicts
- ✅ **Air integration**: Auto-reload on file changes
- ✅ **MCP binary updates**: Keeps VS Code MCP server in sync
- ✅ **Clear instructions**: Shows how to restart VS Code MCP

### **Option 2: VS Code Tasks**

1. Open Command Palette (`Cmd+Shift+P`)
2. Type "Tasks: Run Task"
3. Choose "Dev: Start Auto-Reload Server"

**Features:**
- ✅ **Integrated**: Runs within VS Code
- ✅ **Background process**: Doesn't block VS Code
- ✅ **Output panel**: Shows build logs in VS Code

### **Option 3: Manual Air**

```bash
# Install air if not present
go install github.com/air-verse/air@latest

# Start air directly
~/go/bin/air
```

### **Option 4: Traditional Make Commands**

```bash
make dev        # Uses air or falls back to watch
make watch      # Uses entr for file watching  
make dev-simple # No auto-reload, manual restart
```

## 📋 VS Code Tasks Available

| **Task** | **Purpose** | **Mode** |
|----------|-------------|----------|
| `Dev: Start Auto-Reload Server` | Primary development | HTTP-only |
| `Dev: Stop Server` | Stop all servers | All |
| `Build and Restart All Servers` | Build + restart instructions | All |
| `Build Server` | Build binary only | All |
| `Run Tests` | Execute test suite | N/A |
| `Run Tests with Coverage` | Tests with coverage | N/A |
| `Watch Tests` | Auto-run tests on changes | N/A |

## 🔄 MCP Server Restart Process

### **Automatic (HTTP Server)**
- ✅ **File changes detected** → Air rebuilds → HTTP server restarts
- ✅ **No manual intervention** required
- ✅ **Fast feedback loop** (~1-2 seconds)

### **Manual (VS Code MCP Server)**
After code changes and build:

1. **Press** `Cmd+Shift+P`
2. **Type** "MCP: Restart All Servers"
3. **Press** Enter

**Or alternatively:**
1. **Press** `Cmd+Shift+P`
2. **Type** "Developer: Reload Window"
3. **Press** Enter

## 🚨 Development Considerations

### **⚠️ Port Conflicts Prevention**
- **Problem**: Multiple servers trying to use port `:8080`
- **Solution**: Use mode-specific scripts that cleanup before starting
- **Best Practice**: Always use `./scripts/dev-start.sh` or `./scripts/dev-stop.sh`

### **🔄 Dual Server Management**
- **HTTP Server**: For API testing and development (auto-restarts)
- **STDIO Server**: For VS Code MCP integration (manual restart)
- **Binary Sync**: Air builds both `./tmp/main` and `./bin/mcp-server`

### **📁 File Watching Scope**
Air watches these file types:
```text
✅ .go files (source code)
✅ .json files (API specs, configs)
✅ .env files (environment variables)
❌ .md files (documentation)
❌ Binary files
❌ .git directory
```

### **🎯 Performance Considerations**
- **Fast builds**: Air uses incremental compilation
- **Resource usage**: HTTP server uses minimal resources
- **Memory**: STDIO server loads OpenAPI spec in memory
- **Startup time**: HTTP ~1s, STDIO ~2s (loads full API spec)

### **🐛 Common Issues**

#### **"Address already in use" Error**
```bash
# Solution: Stop all servers first
./scripts/dev-stop.sh
# Then start fresh
./scripts/dev-start.sh
```

#### **Air not found**
```bash
# Install air
go install github.com/air-verse/air@latest
# Add to PATH if needed
export PATH="$PATH:$(go env GOPATH)/bin"
```

#### **VS Code MCP not connecting**
1. Check `.vscode/mcp.json` configuration
2. Restart MCP servers: `Cmd+Shift+P` → "MCP: Restart All Servers"
3. Check VS Code output panel for errors
4. Ensure binary exists: `ls -la bin/mcp-server`

#### **Build Failures**
```bash
# Clean and rebuild
make clean
make build
# Check Go version
go version  # Should be 1.19+
# Update dependencies
go mod tidy
```

## 🐛 Debugging

### **VS Code Debugging**

1. Set breakpoints in your Go code
2. Press `F5` or use Debug panel
3. Select "Debug MCP Server" configuration

### **Manual Debugging with Delve**

```bash
# Install delve if not present
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug built binary
make build
dlv exec ./bin/mcp-server -- --mode http --env .env

# Debug source directly
dlv debug ./cmd/main.go -- --mode stdio --env .env
```

### **HTTP API Testing**

```bash
# Test tools endpoint
curl http://localhost:8080/tools | jq

# Test invoke endpoint
curl -X POST http://localhost:8080/invoke \
  -H "Content-Type: application/json" \
  -d '{"tool": "list", "arguments": {"resource": "environments"}}'
```

## 📁 File Structure for Development

```text
mcp-server/
├── .air.toml                 # Air configuration (HTTP mode)
├── .vscode/
│   ├── tasks.json           # VS Code development tasks
│   ├── launch.json          # Debug configurations
│   ├── settings.json        # MCP server config
│   └── mcp.json            # MCP protocol config
├── scripts/
│   ├── dev-start.sh        # Smart development start
│   ├── dev-stop.sh         # Clean server shutdown
│   └── restart-mcp.sh      # MCP restart helper
├── tmp/                    # Air build cache (auto-created)
│   └── main               # HTTP development binary
├── bin/
│   └── mcp-server         # Production binary (VS Code MCP)
├── internal/              # Go source code
├── cmd/main.go           # Application entry point
└── .env                  # Environment configuration
```

## 💡 Development Tips

### **Efficient Workflow**

1. **Single Terminal**: Use `./scripts/dev-start.sh` for everything
2. **Watch Logs**: Keep terminal visible to see build status
3. **Quick Testing**: Use HTTP API for rapid iteration
4. **VS Code Integration**: Test MCP integration periodically
5. **Clean Restart**: Use stop script when things get stuck

### **Code Changes Best Practices**

1. **Small iterations**: Make small changes, test frequently
2. **Check both modes**: Test HTTP API and VS Code MCP
3. **Environment sync**: Restart VS Code MCP after `.env` changes
4. **Build validation**: Watch for build errors in Air output
5. **Port awareness**: Always stop servers before switching modes

### **Testing Strategy**

```bash
# 1. Unit tests (fast feedback)
make test

# 2. HTTP API testing (integration)
curl http://localhost:8080/tools

# 3. VS Code MCP testing (full integration)
# Use VS Code Chat with MCP commands

# 4. Cross-mode testing
# Test same functionality in both HTTP and STDIO modes
```

### **Performance Monitoring**

- **Build time**: Air shows build duration
- **Memory usage**: Monitor with `top` or Activity Monitor
- **API response**: Use `curl` with timing
- **Startup time**: Check logs for initialization duration

## 🔧 Server Management Commands

```bash
# Development lifecycle
./scripts/dev-start.sh    # Start development (recommended)
./scripts/dev-stop.sh     # Stop all servers
make build-mcp           # Build + show MCP restart instructions

# Traditional make commands
make build               # Build binary only
make clean              # Clean build artifacts
make test               # Run tests
make test-coverage      # Tests with coverage report

# Direct server control
./bin/mcp-server --mode http --env .env     # HTTP only
./bin/mcp-server --mode stdio --env .env    # STDIO only
./bin/mcp-server --mode both --env .env     # Both (conflicts possible)
```

## 🚨 Troubleshooting Quick Reference

| **Problem** | **Solution** |
|-------------|--------------|
| Port 8080 in use | Run `./scripts/dev-stop.sh` |
| Air not found | Run `go install github.com/air-verse/air@latest` |
| Build errors | Run `make clean && make build` |
| VS Code MCP not working | Press `Cmd+Shift+P` → "MCP: Restart All Servers" |
| Environment issues | Check `.env` file exists and has valid credentials |
| Slow builds | Check if `tmp/` directory has too many files |
| Memory issues | Restart development server periodically |

## ✅ Verification Checklist

Before committing changes, verify:

- [ ] **HTTP API works**: `curl http://localhost:8080/tools`
- [ ] **STDIO mode works**: `echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/mcp-server --mode stdio --env .env`
- [ ] **Tests pass**: `make test`
- [ ] **No build errors**: Clean Air output
- [ ] **VS Code MCP connects**: Check MCP server status in VS Code
- [ ] **Environment loaded**: Verify API credentials work
- [ ] **Both binaries updated**: Check timestamps on `tmp/main` and `bin/mcp-server`
