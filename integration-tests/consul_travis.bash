#!/bin/bash

consul_host=http://localhost:8500
root_key=integration-tests
playbook_key=${root_key}/good-postgres.yml

## Load all key-value pairs into Consul
curl -X PUT -d "`cat ${playbook_key}`" ${consul_host}/v1/kv/${playbook_key}
curl -X PUT -d "`cat ${root_key}/postgres-sql/bad/1.sql`" ${consul_host}/v1/kv/${root_key}/postgres-sql/bad/1.sql
curl -X PUT -d "`cat ${root_key}/postgres-sql/good/1.sql`" ${consul_host}/v1/kv/${root_key}/postgres-sql/good/1.sql
curl -X PUT -d "`cat ${root_key}/postgres-sql/good/2a.sql`" ${consul_host}/v1/kv/${root_key}/postgres-sql/good/2a.sql
curl -X PUT -d "`cat ${root_key}/postgres-sql/good/2b.sql`" ${consul_host}/v1/kv/${root_key}/postgres-sql/good/2b.sql
curl -X PUT -d "`cat ${root_key}/postgres-sql/good/3.sql`" ${consul_host}/v1/kv/${root_key}/postgres-sql/good/3.sql
curl -X PUT -d "`cat ${root_key}/postgres-sql/good/assert.sql`" ${consul_host}/v1/kv/${root_key}/postgres-sql/good/assert.sql
