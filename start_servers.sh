#!/usr/bin/env bash
cd /home/ed/janitarr
bun run start > backend.log 2>&1 &
BACKEND_PID=$!
echo "Backend server started with PID: $BACKEND_PID"

cd /home/ed/janitarr/ui
bun run dev > ../frontend.log 2>&1 &
FRONTEND_PID=$!
echo "Frontend server started with PID: $FRONTEND_PID"

echo "Servers are starting in the background. Check backend.log and frontend.log for status."
