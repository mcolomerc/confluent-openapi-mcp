#!/bin/bash

# MCP Server Monitoring Stack Quick Start

set -e

echo "🚀 Starting MCP Server Monitoring Stack..."
echo

# Function to check if docker-compose is available
check_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        echo "✅ Using docker-compose"
        COMPOSE_CMD="docker-compose"
    elif docker compose version &> /dev/null; then
        echo "✅ Using docker compose"
        COMPOSE_CMD="docker compose"
    else
        echo "❌ Docker Compose not found. Please install Docker Compose."
        exit 1
    fi
}

# Function to start the monitoring stack
start_monitoring() {
    echo "📊 Starting monitoring stack..."
    $COMPOSE_CMD -f docker-compose.monitoring.yml up -d
    
    echo
    echo "⏳ Waiting for services to be ready..."
    sleep 10
    
    # Check service health
    echo "🔍 Checking service health..."
    
    # Check MCP Server
    if curl -s http://localhost:8080/health > /dev/null; then
        echo "✅ MCP Server is healthy"
    else
        echo "⚠️  MCP Server not ready yet (this is normal, may take a few more seconds)"
    fi
    
    # Check Prometheus
    if curl -s http://localhost:9090/-/healthy > /dev/null; then
        echo "✅ Prometheus is healthy"
    else
        echo "❌ Prometheus not responding"
    fi
    
    # Check Grafana
    if curl -s http://localhost:3000/api/health > /dev/null; then
        echo "✅ Grafana is healthy"
    else
        echo "❌ Grafana not responding"
    fi
}

# Function to show access information
show_access_info() {
    echo
    echo "🎉 Monitoring stack is running!"
    echo
    echo "📊 Service URLs:"
    echo "  • MCP Server:     http://localhost:8080"
    echo "  • Prometheus:     http://localhost:9090"
    echo "  • Grafana:        http://localhost:3000"
    echo "  • Node Exporter:  http://localhost:9100"
    echo
    echo "🔗 MCP Server Endpoints:"
    echo "  • Health:         http://localhost:8080/health"
    echo "  • JSON Metrics:   http://localhost:8080/metrics"
    echo "  • Prometheus:     http://localhost:8080/metrics/prometheus"
    echo "  • Force GC:       curl -X POST http://localhost:8080/gc"
    echo
    echo "🔐 Grafana Login:"
    echo "  • Username: admin"
    echo "  • Password: admin123"
    echo
    echo "📈 Pre-configured Dashboard:"
    echo "  • Go to Grafana → Dashboards → MCP Server Monitoring"
    echo
}

# Function to stop the monitoring stack
stop_monitoring() {
    echo "🛑 Stopping monitoring stack..."
    $COMPOSE_CMD -f docker-compose.monitoring.yml down
    echo "✅ Monitoring stack stopped"
}

# Function to show logs
show_logs() {
    echo "📋 Showing logs for all services..."
    $COMPOSE_CMD -f docker-compose.monitoring.yml logs -f
}

# Function to show quick test commands
show_test_commands() {
    echo "🧪 Quick test commands:"
    echo
    echo "# Test MCP Server health"
    echo "curl http://localhost:8080/health | jq"
    echo
    echo "# Get Prometheus metrics"
    echo "curl http://localhost:8080/metrics/prometheus"
    echo
    echo "# Force garbage collection"
    echo "curl -X POST http://localhost:8080/gc | jq"
    echo
    echo "# Query Prometheus directly"
    echo "curl 'http://localhost:9090/api/v1/query?query=mcp_memory_alloc_bytes'"
    echo
}

# Main script logic
case "${1:-start}" in
    "start")
        check_docker_compose
        start_monitoring
        show_access_info
        ;;
    "stop")
        check_docker_compose
        stop_monitoring
        ;;
    "restart")
        check_docker_compose
        stop_monitoring
        sleep 2
        start_monitoring
        show_access_info
        ;;
    "logs")
        check_docker_compose
        show_logs
        ;;
    "test")
        show_test_commands
        ;;
    "status")
        check_docker_compose
        $COMPOSE_CMD -f docker-compose.monitoring.yml ps
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|logs|test|status}"
        echo
        echo "Commands:"
        echo "  start    - Start the monitoring stack"
        echo "  stop     - Stop the monitoring stack"
        echo "  restart  - Restart the monitoring stack"
        echo "  logs     - Show logs from all services"
        echo "  test     - Show test commands"
        echo "  status   - Show service status"
        echo
        echo "Quick start:"
        echo "  $0 start"
        echo
        exit 1
        ;;
esac
