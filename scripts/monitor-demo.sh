#!/bin/bash

# Resource Monitoring Demo Script for MCP Server

echo "=== MCP Server Resource Monitoring Demo ==="
echo

# Function to check if server is running
check_server() {
    if ! pgrep -f "mcp-server" > /dev/null; then
        echo "❌ MCP Server is not running"
        return 1
    fi
    echo "✅ MCP Server is running"
    return 0
}

# Function to test HTTP endpoints
test_endpoints() {
    echo "Testing monitoring endpoints..."
    
    # Test health endpoint
    echo "📊 Health endpoint:"
    curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "Health endpoint not responding"
    echo
    
    # Test metrics endpoint
    echo "📈 Metrics endpoint:"
    curl -s http://localhost:8080/metrics | jq '.' 2>/dev/null || echo "Metrics endpoint not responding"
    echo
    
    # Test GC endpoint
    echo "🗑️  Triggering garbage collection:"
    curl -s -X POST http://localhost:8080/gc | jq '.' 2>/dev/null || echo "GC endpoint not responding"
    echo
}

# Function to demonstrate different monitoring modes
demo_monitoring_modes() {
    echo "=== Monitoring Mode Examples ==="
    echo
    
    echo "1. Start server with monitoring every 10 seconds:"
    echo "   ./bin/mcp-server -mode=http -monitor=10s"
    echo
    
    echo "2. Start server with monitoring every 1 minute:"
    echo "   ./bin/mcp-server -mode=http -monitor=1m"
    echo
    
    echo "3. Start server with monitoring disabled:"
    echo "   ./bin/mcp-server -mode=http -monitor=off"
    echo
    
    echo "4. Start server with both stdio and http (default monitoring):"
    echo "   ./bin/mcp-server -mode=both"
    echo
}

# Function to show available endpoints
show_endpoints() {
    echo "=== Available Monitoring Endpoints ==="
    echo
    echo "🔗 HTTP Endpoints (when server runs in http or both mode):"
    echo "   GET  /health  - Quick health check with basic metrics"
    echo "   GET  /metrics - Comprehensive resource metrics"
    echo "   POST /gc      - Force garbage collection and show before/after metrics"
    echo "   GET  /mcp     - Main MCP protocol endpoint"
    echo
}

# Function to show resource monitoring tips
show_tips() {
    echo "=== Resource Monitoring Tips ==="
    echo
    echo "💡 Console Monitoring:"
    echo "   - Metrics are logged to stderr at specified intervals"
    echo "   - Default interval is 30 seconds"
    echo "   - Use -monitor=off to disable console logging"
    echo
    echo "🌐 HTTP Monitoring:"
    echo "   - Real-time metrics available via HTTP endpoints"
    echo "   - Use curl, wget, or any HTTP client to fetch metrics"
    echo "   - JSON format for easy integration with monitoring tools"
    echo
    echo "🔧 Garbage Collection:"
    echo "   - Use /gc endpoint to manually trigger GC"
    echo "   - Shows memory usage before and after collection"
    echo "   - Useful for debugging memory issues"
    echo
    echo "📊 Metric Types:"
    echo "   - Memory: Allocated, System, Heap, Stack usage"
    echo "   - CPU: Number of CPUs, CGO calls"
    echo "   - Concurrency: Number of goroutines"
    echo "   - GC: Garbage collection statistics"
    echo
}

# Main execution
case "${1:-demo}" in
    "test")
        if check_server; then
            test_endpoints
        else
            echo "Please start the server first:"
            echo "  ./bin/mcp-server -mode=http -monitor=10s"
        fi
        ;;
    "endpoints")
        show_endpoints
        ;;
    "tips")
        show_tips
        ;;
    "modes")
        demo_monitoring_modes
        ;;
    *)
        echo "Usage: $0 [test|endpoints|tips|modes]"
        echo
        demo_monitoring_modes
        show_endpoints
        show_tips
        
        echo "=== Quick Start ==="
        echo "1. Start the server with monitoring:"
        echo "   ./bin/mcp-server -mode=http -monitor=10s"
        echo
        echo "2. In another terminal, test the endpoints:"
        echo "   $0 test"
        echo
        ;;
esac
