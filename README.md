# ssh-exec-demo

## GIST

- https://gist.github.com/tamalsaha/8ad260a679bcd65d21d52fa709171af3

## Step 1: Listen for Logs

```
nats -s this-is-nats.appscode.ninja --user=$THIS_IS_NATS_USERNAME --password=$THIS_IS_NATS_PASSWORD sub stackscript-log
```

## Step 2: Run script

```
./script.sh
```

## Send Logs via NATS (used in stackscript)

```
SHIPPER_FILE=/tmp/stackscript.log \
  SHIPPER_SUBJECT=stackscript-log \
  go run producer/main.go
```

## NATS example

```
nats -s this-is-nats.appscode.ninja --user=$THIS_IS_NATS_USERNAME --password=$THIS_IS_NATS_PASSWORD sub stackscript-log

nats -s this-is-nats.appscode.ninja --user=$THIS_IS_NATS_USERNAME --password=$THIS_IS_NATS_PASSWORD pub hello tamal
```
