#!/bin/bash
set -euo pipefail

# Configuration
BACKUP_DIR="/home/mier/dev/url/backups"
CONTAINER_NAME="url-postgres"
DB_USER="urlshortener"
DB_NAME="url"
KEEP_DAYS=7

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/url_backup_$TIMESTAMP.sql.gz"

echo "[$(date)] Starting backup..."

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "[$(date)] ERROR: Container $CONTAINER_NAME is not running"
    exit 1
fi

# Create backup
if docker exec "$CONTAINER_NAME" pg_dump -U "$DB_USER" -d "$DB_NAME" | gzip > "$BACKUP_FILE"; then
    FILESIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "[$(date)] Backup created: $BACKUP_FILE ($FILESIZE)"
else
    echo "[$(date)] ERROR: Backup failed"
    rm -f "$BACKUP_FILE"
    exit 1
fi

# Verify backup is not empty
if [ ! -s "$BACKUP_FILE" ]; then
    echo "[$(date)] ERROR: Backup file is empty"
    rm -f "$BACKUP_FILE"
    exit 1
fi

# Rotate old backups
echo "[$(date)] Removing backups older than $KEEP_DAYS days..."
find "$BACKUP_DIR" -name "url_backup_*.sql.gz" -mtime +$KEEP_DAYS -delete

# Show current backups
echo "[$(date)] Current backups:"
ls -lh "$BACKUP_DIR"/url_backup_*.sql.gz 2>/dev/null || echo "  (none)"

echo "[$(date)] Backup completed successfully"
