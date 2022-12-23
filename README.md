# ssh-exec-demo

## Step 1: Listen for Logs

```
nats -s this-is-nats.appscode.ninja --user=$THIS_IS_NATS_USERNAME --password=$THIS_IS_NATS_PASSWORD sub stackscript-log
```

## Step 2: Send Logs

```
SHIPPER_FILE=/tmp/stackscript.log \
  SHIPPER_SUBJECT=stackscript-log \
  go run producer/main.go
```

## Step 3: Run script

```
./script.sh
```

## NATS example

```
nats -s this-is-nats.appscode.ninja --user=$THIS_IS_NATS_USERNAME --password=$THIS_IS_NATS_PASSWORD sub stackscript-log

nats -s this-is-nats.appscode.ninja --user=$THIS_IS_NATS_USERNAME --password=$THIS_IS_NATS_PASSWORD pub hello tamal
```
