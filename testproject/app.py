import os

db_host = os.getenv("DB_HOST")
db_name = os.environ["DB_NAME"]
app_port = os.environ.get("APP_PORT")
debug = os.getenv("DEBUG_MODE")
unknown = os.getenv("UNKNOWN_VAR")