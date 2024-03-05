# env2cfg
env2cfg substitutes variables in specified template with variables found in a dotenv file or environment.

## Usage
```bash
Usage: env2cfg [OPTIONS] <env-file> <template-file> [<output-file>]

Substitutes variables in template with values found in environment and/or env file.

Arguments:
  <env-file>         File to read variables from or - to use only environment.
  <template-file>    Template file to use.
  [<output-file>]    File to output completed template or stdout if omitted.

Flags:
  -h, --help                      Show context-sensitive help.
  -D, --dry-run                   Don't write to output-file.
  -E, --env-override              Favor envfile over environment.
  -e, --env-vars=KEY=VALUE;...    Set environment variables from command line.
  -P, --no-env-perms              Don't check env-file file permissions.
  -p, --output-perms="0640"       Set output-file permissions.
  -q, --quiet                     Don't display warnings.
  -v, --version                   Show version
      --delim-start="${"          Starting delimiter string.
      --delim-end="}"             Ending delimiter string.
```
