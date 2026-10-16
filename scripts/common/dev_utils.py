import subprocess
from common.utils import copy_assets, create_snapshot, restore_snapshot, run_server

def setup_environment():
    """Prepare directories, snapshot DB, and copy assets."""
    print("📸 Step 1: Creating database snapshot...")
    create_snapshot()
    
    print("📦 Step 2: Copying assets...")
    copy_assets()
    print("✅ Environment ready.")

def cleanup_and_restore(process):
    """Stop the server and restore the database snapshot."""
    print("\n🛑 Stopping server...")
    if process:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()

    print("🔄 Step 3: Rolling back database...")
    restore_snapshot()
    print("✅ Development session ended. Database restored.")