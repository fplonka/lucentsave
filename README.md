setup db with:

```
export LS2_DB_PASSWORD='pass_here'
psql -U postgres -f init_db.sql

```

or actually:
sudo -u postgres psql -U postgres -f init_db.sql
??

on vps had to also modify hba file ?? to set trust everywhere instead of peer

and port on vps is 5433 because 5432 was for postgres 14 which i nuked

needs pgvector installed

use lslog (alias for tail -f src/log.txt | jq '.') to pretty print recent logs

TODO:
- inline js/css? in base template