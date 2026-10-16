import os
import sys
import subprocess
import shutil

# Resolve PROJECT_ROOT from Makefile env, or fallback to directory math
PROJECT_ROOT = os.getenv("PROJECT_ROOT")
if not PROJECT_ROOT:
    PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from common.dump import dump_database, load_database

# --- Paths ---
BACKUP_DIR = os.path.join(PROJECT_ROOT, "backup")
SNAPSHOT_FILE = os.path.join(BACKUP_DIR, "snapshot.sql")
ASSET_SRC = os.path.join(PROJECT_ROOT, "asset")
ASSET_DST = os.path.join(PROJECT_ROOT, "build", "asset")
IS_WINDOWS = os.name == 'nt'
SERVER_BINARY = os.path.join(PROJECT_ROOT, "build", "main")

def copy_assets():
    """Copy assets from source to build directory."""
    if not os.path.exists(ASSET_SRC): return
    if os.path.exists(ASSET_DST): shutil.rmtree(ASSET_DST)
    shutil.copytree(ASSET_SRC, ASSET_DST)

def create_snapshot():
    """Create a database snapshot."""
    os.makedirs(BACKUP_DIR, exist_ok=True)
    dump_database(SNAPSHOT_FILE)

def restore_snapshot():
    """Restore database from the latest snapshot."""
    if os.path.isfile(SNAPSHOT_FILE):
        load_database(SNAPSHOT_FILE)

def run_server():
    """Start the Go server and return the process object."""
    if not os.path.isfile(SERVER_BINARY):
        raise FileNotFoundError(f"Server binary not found at {SERVER_BINARY}. Run 'make build' first.")
    
    print("🏃 Starting server... (Press Ctrl+C to stop and rollback)")
    return subprocess.Popen([SERVER_BINARY], cwd=PROJECT_ROOT)