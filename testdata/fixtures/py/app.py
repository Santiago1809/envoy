import os

db_host = os.environ['DB_HOST']
db_port = os.getenv('DB_PORT')
api_key = os.environ.get('API_KEY')

config = {
    'database': f'postgres://{db_host}:{db_port}',
    'api_key': api_key
}

print(config)
