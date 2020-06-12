#DDLog


DataDog log tool. Simple CLI tool for getting logs/stats etc.

##Usage

Firstly, create API/APP key from within the Datadog site and populate config.json

There are a number of assumptions (will be removed/configured later). You have a facet called
"environment" used to indicate the environment that the log belongs. For me, it's test,stage or prod.

A second assumption is that you have a facet called "level" which is info,warn or error.

##CLI Options

- env : test/stage/prod
- levels : info/warn/error.  Can be comma delimited to include multiple. eg. -levels "warn,error"
- query : Raw Datadog query that is appended to the built in query (filtering via env and level)
- mins : Last N minutes of logs to be searched
- stats : Just show stats (counts currently) instead of displaying entire logs.




