#!/usr/bin/env python3
"""
PostgreSQL Database Connection Script
A script to connect to the PostgreSQL database using the provided connection string.
"""

import psycopg2
import sys
from urllib.parse import urlparse
from typing import Optional, Dict, Any, List
import json
from datetime import datetime

class DatabaseConnection:
    def __init__(self, connection_string: str):
        """
        Initialize the database connection.
        
        Args:
            connection_string (str): PostgreSQL connection string
        """
        self.connection_string = connection_string
        self.connection = None
        self.cursor = None
        
    def connect(self) -> bool:
        """
        Establish connection to the PostgreSQL database.
        
        Returns:
            bool: True if connection successful, False otherwise
        """
        try:
            # Parse the connection string
            parsed_url = urlparse(self.connection_string)
            
            # Extract connection parameters
            self.connection = psycopg2.connect(
                host=parsed_url.hostname,
                port=parsed_url.port,
                database=parsed_url.path[1:],  # Remove leading '/'
                user=parsed_url.username,
                password=parsed_url.password
            )
            
            self.cursor = self.connection.cursor()
            print("✅ Successfully connected to PostgreSQL database!")
            return True
            
        except psycopg2.Error as e:
            print(f"❌ Error connecting to database: {e}")
            return False
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
            return False
    
    def disconnect(self):
        """Close the database connection."""
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()
        print("🔌 Database connection closed.")
    
    def execute_query(self, query: str, params: tuple = None) -> Optional[List[Dict[str, Any]]]:
        """
        Execute a SQL query and return results.
        
        Args:
            query (str): SQL query to execute
            params (tuple): Query parameters (optional)
            
        Returns:
            Optional[List[Dict[str, Any]]]: Query results as list of dictionaries
        """
        if not self.connection or not self.cursor:
            print("❌ No active database connection!")
            return None
            
        try:
            self.cursor.execute(query, params)
            
            # If it's a SELECT query, fetch results
            if query.strip().upper().startswith('SELECT'):
                columns = [desc[0] for desc in self.cursor.description]
                rows = self.cursor.fetchall()
                
                # Convert to list of dictionaries
                results = []
                for row in rows:
                    results.append(dict(zip(columns, row)))
                
                return results
            else:
                # For INSERT, UPDATE, DELETE queries
                self.connection.commit()
                return [{"affected_rows": self.cursor.rowcount}]
                
        except psycopg2.Error as e:
            print(f"❌ Database error: {e}")
            self.connection.rollback()
            return None
        except Exception as e:
            print(f"❌ Unexpected error: {e}")
            return None
    
    def get_table_info(self, table_name: str) -> Optional[List[Dict[str, Any]]]:
        """
        Get information about a specific table.
        
        Args:
            table_name (str): Name of the table
            
        Returns:
            Optional[List[Dict[str, Any]]]: Table information
        """
        query = """
        SELECT 
            column_name,
            data_type,
            is_nullable,
            column_default
        FROM information_schema.columns 
        WHERE table_name = %s
        ORDER BY ordinal_position;
        """
        return self.execute_query(query, (table_name,))
    
    def list_tables(self) -> Optional[List[Dict[str, Any]]]:
        """
        List all tables in the database.
        
        Returns:
            Optional[List[Dict[str, Any]]]: List of tables
        """
        query = """
        SELECT 
            table_name,
            table_type
        FROM information_schema.tables 
        WHERE table_schema = 'public'
        ORDER BY table_name;
        """
        return self.execute_query(query)
    
    def get_database_info(self) -> Optional[Dict[str, Any]]:
        """
        Get general database information.
        
        Returns:
            Optional[Dict[str, Any]]: Database information
        """
        info = {}
        
        # Get database name
        db_name_query = "SELECT current_database();"
        db_name_result = self.execute_query(db_name_query)
        if db_name_result:
            info['database_name'] = db_name_result[0]['current_database']
        
        # Get PostgreSQL version
        version_query = "SELECT version();"
        version_result = self.execute_query(version_query)
        if version_result:
            info['postgresql_version'] = version_result[0]['version']
        
        # Get current user
        user_query = "SELECT current_user;"
        user_result = self.execute_query(user_query)
        if user_result:
            info['current_user'] = user_result[0]['current_user']
        
        return info

def main():
    """
    Main function to demonstrate database connection and operations.
    """
    # Your database connection string
    CONNECTION_STRING = "postgres://postgres:Badho_1301@db.badho.in:5432/badho-app"
    
    # Create database connection instance
    db = DatabaseConnection(CONNECTION_STRING)
    
    try:
        # Connect to database
        if not db.connect():
            print("❌ Failed to connect to database. Exiting.")
            sys.exit(1)
        
        print("\n" + "="*60)
        print("🗄️  DATABASE CONNECTION SUCCESSFUL")
        print("="*60)
        
        # Get database information
        print("\n📊 Database Information:")
        db_info = db.get_database_info()
        if db_info:
            for key, value in db_info.items():
                print(f"   {key}: {value}")
        
        # List tables
        print("\n📋 Available Tables:")
        tables = db.list_tables()
        if tables:
            for table in tables:
                print(f"   - {table['table_name']} ({table['table_type']})")
        else:
            print("   No tables found or error retrieving tables.")
        
        # Interactive mode
        print("\n" + "="*60)
        print("🔍 INTERACTIVE DATABASE QUERY MODE")
        print("="*60)
        print("Type SQL queries to execute (type 'quit' to exit)")
        print("Examples:")
        print("  - SELECT * FROM users LIMIT 5;")
        print("  - SELECT COUNT(*) FROM products;")
        print("  - \\d table_name (to describe a table)")
        print("="*60)
        
        while True:
            try:
                query = input("\n💻 SQL> ").strip()
                
                if query.lower() in ['quit', 'exit', 'q']:
                    break
                
                if query.lower().startswith('\\d'):
                    # Handle table description
                    table_name = query.split()[1] if len(query.split()) > 1 else None
                    if table_name:
                        print(f"\n📋 Table structure for '{table_name}':")
                        table_info = db.get_table_info(table_name)
                        if table_info:
                            print(f"{'Column':<20} {'Type':<15} {'Nullable':<10} {'Default'}")
                            print("-" * 60)
                            for col in table_info:
                                print(f"{col['column_name']:<20} {col['data_type']:<15} {col['is_nullable']:<10} {col['column_default'] or 'None'}")
                        else:
                            print(f"❌ Table '{table_name}' not found or error occurred.")
                    else:
                        print("❌ Please specify table name: \\d table_name")
                    continue
                
                if not query:
                    continue
                
                # Execute the query
                results = db.execute_query(query)
                
                if results is not None:
                    if results:
                        print(f"\n✅ Query executed successfully. {len(results)} row(s) returned.")
                        
                        # Display results in a nice format
                        if len(results) > 0:
                            print("\n📊 Results:")
                            print("-" * 80)
                            
                            # Print column headers
                            if isinstance(results[0], dict):
                                columns = list(results[0].keys())
                                header = " | ".join(f"{col:<15}" for col in columns)
                                print(header)
                                print("-" * 80)
                                
                                # Print rows (limit to 10 for display)
                                for i, row in enumerate(results[:10]):
                                    row_str = " | ".join(f"{str(row[col]):<15}" for col in columns)
                                    print(row_str)
                                
                                if len(results) > 10:
                                    print(f"... and {len(results) - 10} more rows")
                            else:
                                print(results)
                    else:
                        print("✅ Query executed successfully. No rows returned.")
                else:
                    print("❌ Query failed. Check your SQL syntax.")
                    
            except KeyboardInterrupt:
                print("\n\n👋 Goodbye!")
                break
            except Exception as e:
                print(f"❌ Error: {e}")
    
    except Exception as e:
        print(f"❌ Unexpected error: {e}")
    
    finally:
        # Always close the connection
        db.disconnect()

if __name__ == "__main__":
    main()
