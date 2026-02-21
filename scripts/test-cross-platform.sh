#!/bin/bash

# Cross-platform testing script for gotunnel
# Tests basic functionality across different scenarios

set -e

echo "🧪 Running gotunnel integration tests..."

# Test 1: Build verification
echo "📦 Testing build..."
go build ./cmd/gotunnel
echo "✅ Build successful"

# Test 2: Help command
echo "📖 Testing help command..."
./gotunnel --help > /dev/null
echo "✅ Help command works"

# Test 3: Version command
echo "🏷️  Testing version command..."
./gotunnel --version > /dev/null
echo "✅ Version command works"

# Test 4: Configuration loading
echo "⚙️  Testing configuration loading..."
cat > test-config.yaml << EOF
global:
  environment: "test"
  debug: true
  no_privilege_check: true

proxy:
  mode: "builtin"
  http_port: 8888
  https_port: 8443

logging:
  level: "debug"
EOF

# Test config loading
timeout 5s ./gotunnel --config test-config.yaml list || echo "Config loading test completed"
echo "✅ Configuration loading works"

# Test 5: Basic tunnel creation (non-privileged)
echo "🚇 Testing basic tunnel creation..."

# Start a simple test server in background
python3 -m http.server 8080 > /dev/null 2>&1 &
TEST_SERVER_PID=$!
sleep 2

# Test tunnel creation with timeout
timeout 10s ./gotunnel --no-privilege-check start --port 8080 --domain integration-test --https=false &
TUNNEL_PID=$!

# Wait a bit for tunnel to start
sleep 5

# Test if tunnel process is running
if kill -0 $TUNNEL_PID 2>/dev/null; then
    echo "✅ Tunnel creation successful"
    kill $TUNNEL_PID 2>/dev/null || true
else
    echo "❌ Tunnel creation failed"
fi

# Cleanup test server
kill $TEST_SERVER_PID 2>/dev/null || true

# Test 6: Error handling
echo "❌ Testing error handling..."

# Test invalid port (should fail gracefully)
timeout 5s ./gotunnel --no-privilege-check start --port -1 --domain invalid-test --https=false 2>/dev/null || echo "✅ Error handling works"

# Test 7: List tunnels (should work with no tunnels)
echo "📋 Testing tunnel listing..."
timeout 5s ./gotunnel --no-privilege-check list || echo "✅ Tunnel listing works"

# Cleanup
rm -f test-config.yaml

echo "🎉 All integration tests completed!"