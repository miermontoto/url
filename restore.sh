#!/bin/bash
set -euo pipefail

# Configuration
BACKUP_DIR="/home/mier/dev/url/backups"
CONTAINER_NAME="url-postgres"
DB_USER="urlshortener"
DB_NAME="url"

# Check arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 <backup_file.sql.gz>"
    echo ""
    echo "Available backups:"
    ls -lht "$BACKUP_DIR"/url_backup_*.sql.gz 2>/dev/null || echo "  (none)"
    exit 1
fi

BACKUP_FILE="$1"

# If only filename given, prepend backup directory
if [[ ! "$BACKUP_FILE" == /* ]]; then
    BACKUP_FILE="$BACKUP_DIR/$BACKUP_FILE"
fi

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "ERROR: Container $CONTAINER_NAME is not running"
    exit 1
fi

echo "WARNING: This will overwrite all data in the database!"
echo "Backup file: $BACKUP_FILE"
read -p "Are you sure? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo "Restore cancelled"
    exit 0
fi

echo "[$(date)] Starting restore..."

# Drop and recreate database
docker exec "$CONTAINER_NAME" psql -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
docker exec "$CONTAINER_NAME" psql -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"

# Restore from backup
if gunzip -c "$BACKUP_FILE" | docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1; then
    echo "[$(date)] Restore completed successfully"

    # Show record counts
    echo ""
    echo "Restored data:"
    docker exec "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT 'urls' as table_name, COUNT(*) as count FROM urls UNION ALL SELECT 'users', COUNT(*) FROM users;"
else
    echo "[$(date)] ERROR: Restore failed"
    exit 1
fi
