import os
import psycopg2
from psycopg2 import sql

def get_db_path():
    """Get database path from environment or default."""
    return os.getenv("DB_PATH", "postgres://postgres:postgres@localhost:5432/mugiko?sslmode=disable")

def dump_database(output_path):
    """Create a logical dump of the 'public' schema."""
    db_path = get_db_path()
    output_dir = os.path.dirname(output_path)
    if output_dir and not os.path.exists(output_dir):
        os.makedirs(output_dir)

    print(f"  -> Dumping DB to {output_path}...")
    conn = psycopg2.connect(db_path)
    cursor = conn.cursor()
    
    cursor.execute("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;")
    tables = [row[0] for row in cursor.fetchall()]
    
    with open(output_path, 'w', encoding='utf-8') as f:
        f.write("BEGIN;\nSET session_replication_role = 'replica';\n\n")
        for table in tables:
            cursor.execute(f"SELECT column_name FROM information_schema.columns WHERE table_name = '{table}' ORDER BY ordinal_position;")
            columns = [row[0] for row in cursor.fetchall()]
            if not columns: continue
            
            cursor.execute(sql.SQL("SELECT * FROM {}").format(sql.Identifier(table)))
            for row in cursor.fetchall():
                vals = ["NULL" if v is None else f"'{str(v).replace(chr(39), chr(39)+chr(39))}'" if isinstance(v, str) else str(v) for v in row]
                f.write(f"INSERT INTO \"{table}\" ({', '.join(columns)}) VALUES ({', '.join(vals)});\n")
        f.write("\nSET session_replication_role = 'origin';\nCOMMIT;\n")
            
    cursor.close()
    conn.close()

def load_database(input_path):
    """Load a SQL file into the database."""
    if not os.path.isfile(input_path):
        print(f"  -> Snapshot file not found: {input_path}")
        return

    db_path = get_db_path()
    print(f"  -> Restoring DB from {input_path}...")
    
    conn = psycopg2.connect(db_path)
    cursor = conn.cursor()
    try:
        with open(input_path, "r", encoding="utf-8") as f:
            cursor.execute(f.read())
        conn.commit()
    except Exception as e:
        conn.rollback()
        print(f"  -> Restore failed: {e}")
    finally:
        cursor.close()
        conn.close()