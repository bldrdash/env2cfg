# env2cfg
env2cfg substitutes variables in specified template with variables found in a dotenv file or environment.

## Installation
```bash
go install github.com:bldrdash/env2cfg@latest
```

## Usage
```bash
cba@snow:~/work/env2cfg$ ./env2cfg -H
env2cfg reads environment variables and produces a config file based on a template

Usage: 
  env2cfg [FLAGS] <template> [<dotenv>] [<output>]
  env2cfg -G <template> [<dotenv>]
  env2cfg -H

Details:
env2cfg will read environment variables from the system and/or the <dotenv> file and 
output to <output>.  If <dotenv> is omitted, nly the system environment will be used 
for variables. <output> is optional and will default to stdout if not provided.

When envoked with the -G flag, env2cfg will generate the <dotenv> file based on 
variables found in <template> If <dotenv> is omitted, the output will be written to stdout.

<template> can be in any format and will be parsed for variables using --delim-start 
and --delim-end.  The default delimiters are "${" and "}".

Example Template:
  mqtt:
    broker: tcp://${MQTT_BROKER}:${MQTT_PORT}
    username: ${MQTT_USER}
    password: ${MQTT_PASS}

Flags:
  -D, --dry-run              Don't write to output-file.
  -G, --gen                  Generate <dotenv> based on <template>.
  -E, --override             Favor envfile over environment.
  -e, --vars key=value;...   Add variables from command line.
  -p, --perms string         Set <output> permissions. (default "0640")
  -I, --ignore-perm          Don't check <envfile> file permissions.
  -q, --quiet                Don't display warnings.
  -v, --version              Show version.
  -H, --detailed             Show detailed help and example.
      --delim-start string   Starting delimiter string. (default "${")
      --delim-end string     Ending delimiter string. (default "}")
```
