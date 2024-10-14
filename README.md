# weko-autox ver.1.1.1
weko auto unit test tool

## Description
This script is unofficial tool for running the unit tests of the Weko3 modules.

## Usage
`autox.sh [-n] [-r] [-o output] [-k] [-v] [-h] [all|weko|invenio] [module1 module2 ...]`

### Commands
* `all`:     Run tests for all modules.
* `invenio`: Run tests for all invenio modules.
* `weko`:    Run tests for all weko modules.

### Arguments
`module1 module2 ...` : Specify the module names to run tox optionally.

### Options
* `-n`  specify the module names to do not run tox by arguments.
        Need to specify the module names to run tox.
* `-r`  Remove the egg-info and .tox directories.
        When permission problems occur, use this option.
* `-o`  Specify the output directory for the log files by argument.
* `-k`  Stop the tox process.
* `-v`  Show the version.
* `-h`  Show the help message.

## Note
The log files are stored in the log directory.

> [!IMPORTANT]
> The following conditions must be satisfied in order for the progress o be displayed correctly
> - The display must fit on a single line.
> - docker does not issue a warning. Create a file in the project root as shown below.
> 
>       # .env
>       ELASTICSEARCH_S3_ACCESS_KEY=
>       ELASTICSEARCH_S3_SECRET_KEY=
>       ELASTICSEARCH_S3_ENDPOINT=
>       ELASTICSEARCH_S3_BUCKET=



## Change Log
### ver.1.1.1
delete options: `-i` and `-w`.  
add options: `-o`; specify the output directory for the log files. `-v`; show the version.

### ver.1.1.0
add commands: all, invenio, weko.

### ver.1.0
the first script.
