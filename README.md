## The Stampkeeper 

Stampkeeper can monitor multiple swarm postage stamps.
It will top them up and dilute stamps as required.

## CLI

```
Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  list        List watched stamps this session
  start       Start stampkeeper
  stop        Stop stampkeeper
  unwatch     Stop watching a batch
  watch       Watch a batch

Flags:
      --config string      config file (default is $HOME/.stampkeeper.yaml)
  -h, --help               help for stampkeeper
      --verbosity string   verbosity level (default "5")
```

### start

Start command has `--server` flag, which will point to the bee debug api which will be used to fund the stamps.

```
   $ stampkeeper start --server http://localhost:1635
```

### watch 

Watch command has the following flags

```
Flags:
      --batch string      BatchId to topup
  -h, --help              help for watch
      --interval string   Interval to check for balance (default "30s")
      --min string        Minimum balance for topup (default "2000000")
      --name string       Custom identifier
      --top string        Amount to be topped up (default "5000000")
      --url string        Bee-Debug-Api Endpoint to check stamp balance
      
Example:

    $ stampkeeper watch --name "my shiny new batch" --batch 6a50032864056992563cee7e31b3323bd25ac34c157f658d02b32a59e241f7f3 --url http://localhost:1635 
```

### unwatch

Unwatch command has the following flags

```
Flags:
      --batch string      BatchId to unwatch
  -h, --help              help for watch
      
Example:

    $ stampkeeper unwatch --batch 6a50032864056992563cee7e31b3323bd25ac34c157f658d02b32a59e241f7f3 
```

### list

List command will list all the stamps watched this session

```
Example:

    $ stampkeeper list 
    
Output:

[
    {
            "active": true,
            "batch": "6a50032864056992563cee7e31b3323bd25ac34c157f658d02b32a59e241f7f3"
    },
    {
            "active": true,
            "batch": "20d6a5ab78177e878ac7cc7c88969ac24c8438bef9347320b16cd69dec6164dd"
    }
]

```

## Bot

stampkeeper only supports telegram bot as of now. To enable bot add `telegram_bot_token` in the config file

It has the following commands
```
/version - Stampkeeper version
/list - List of watched stamps
/watch - Watch a batch

	/watch customName batchId balanceEndpoint minBalance topupBalance interval
	
	Please choose a name without spaces

/unwatch - Stop watching a batch
	
	/unwatch batchId

/help - General usage instruction
```

## Activity Log
Stampkeeper appends all batches in .stampkeeper.yaml and transactions in stampkeeper_accountant.json
