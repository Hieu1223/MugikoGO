#!/usr/bin/env python3
"""
Main entry point for the development server.
"""
import sys
import os

# Ensure the 'scripts' directory is in the Python path so 'common' can be imported
SCRIPTS_DIR = os.path.dirname(os.path.abspath(__file__))
if SCRIPTS_DIR not in sys.path:
    sys.path.insert(0, SCRIPTS_DIR)

from common.dev_utils import setup_environment, cleanup_and_restore
from common.utils import run_server

def main():
    print("🚀 Starting Development Environment...")
    server_process = None
    
    try:
        setup_environment()
        server_process = run_server()
        server_process.wait()  # Block until server exits
        
    except KeyboardInterrupt:
        pass  # Handled gracefully in the finally block
    except Exception as e:
        print(f"\n❌ Fatal Error: {e}")
    finally:
        # This ALWAYS runs, ensuring the DB is restored even on crashes
        cleanup_and_restore(server_process)

if __name__ == "__main__":
    main()