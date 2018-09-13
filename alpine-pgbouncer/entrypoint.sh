#!/bin/sh

# Here are some parameters. See all on
# https://pgbouncer.github.io/config.html

PG_LOG=/var/log/pgbouncer
PG_CONFIG_DIR=/etc/pgbouncer
PG_USER=postgres

# Come from here:
#   https://stackoverflow.com/a/19347380/2431728
#
ESCAPED_DB_USER=${DB_USER//[$'\t\r\n']}
ESCAPED_DB_PASSWORD=${DB_PASSWORD//[$'\t\r\n']}

if [ ! -f ${PG_CONFIG_DIR}/pgbouncer.ini ]; then
  echo "create pgbouncer config in ${PG_CONFIG_DIR}"
  mkdir -p ${PG_CONFIG_DIR}

  printf "\
#pgbouncer.ini
# Description
# Config file is in “ini” format. Section names are between “[” and “]”.
# Lines starting with “;” or “#” are taken as comments and ignored.
# The characters “;” and “#” are not recognized when they appear later in the line.
[pgbouncer]
listen_addr = *
listen_port = 6432
auth_type = any
; When server connection is released back to pool:
;   session      - after client disconnects
;   transaction  - after transaction finishes
;   statement    - after statement finishes
pool_mode = ${POOL_MODE:-transaction}
max_client_conn = ${MAX_CLIENT_CONN:-500}
default_pool_size = ${DEFAULT_POOL_SIZE:-1}
min_pool_size = ${MIN_POOL_SIZE:-0}
max_db_connections = ${MAX_DB_CONNECTIONS:-100}
reserve_pool_size = ${RESERVE_POOL_SIZE:-1}
reserve_pool_timeout = ${RESERVE_POOL_TIMEOUT:-5.0}
server_lifetime = ${SERVER_LIFETIME:-3600}
server_idle_timeout = ${SERVER_IDLE_TIMEOUT:-600}
log_connections = ${LOG_CONNECTIONS:-1}
log_disconnections = ${LOG_DISCONNECTIONS:-1}
log_pooler_errors = ${LOG_POOLER_ERRORS:-1}
stats_period = ${STATS_PERIOD:-60}
ignore_startup_parameters = ${IGNORE_STARTUP_PARAMETERS}
tcp_keepalive = ${TCP_KEEPALIVE:-1}

[databases]
* = host=${DB_HOST:?"Setup pgbouncer config error! You must set DB_HOST env"} port=${DB_PORT:-5432} user=${ESCAPED_DB_USER:-postgres} ${ESCAPED_DB_PASSWORD:+password=${DB_PASSWORD}}
################## end file ##################
" > ${PG_CONFIG_DIR}/pgbouncer.ini
fi

adduser ${PG_USER}
mkdir -p ${PG_LOG}
chmod -R 755 ${PG_LOG}
chown -R ${PG_USER}:${PG_USER} ${PG_LOG}

cat ${PG_CONFIG_DIR}/pgbouncer.ini
echo "Starting pgbouncer..."
exec pgbouncer -u ${PG_USER} ${PG_CONFIG_DIR}/pgbouncer.ini