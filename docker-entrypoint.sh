#!/bin/sh
set -e

PUID=${PUID:-1000}
PGID=${PGID:-1000}

# Create group if it doesn't exist
if ! getent group janitarr > /dev/null 2>&1; then
    addgroup -g "${PGID}" janitarr
fi

# Create user if it doesn't exist
if ! getent passwd janitarr > /dev/null 2>&1; then
    adduser -D -u "${PUID}" -G janitarr -h /data -s /sbin/nologin janitarr
fi

# Ensure correct ownership
chown -R janitarr:janitarr /data

# Drop privileges and execute
exec su-exec janitarr:janitarr "$@"
