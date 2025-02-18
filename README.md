# weko-autox ver.1.2.0
weko auto unit test tool

## Description
This script is unofficial tool for running the unit tests of the Weko3 modules.

## Install
```shell
$ cd weko
$ git clone https://github.com/ivis-kuroda/weko-autox.git auto
$ ln -s auto/autox.sh autox
$ ./autox -v
autox.sh - ver.1.2.0
```

## Usage
`autox.sh [-n] [-r] [-p module] [-o output] [-k] [-v] [-h] [all|weko|invenio] [target1 target2 ...]`

### Commands
* `all`:     Run tests for all modules.
* `invenio`: Run tests for all invenio modules.
* `weko`:    Run tests for all weko modules.

### Arguments
`target1 target2 ...` : Specify the module names to run tox optionally.

### Options
* `-n`  specify the module names to do not run tox by arguments.
* `-r`  Remove the egg-info and .tox directories.
        When permission problems occur, use this option.
* `-p`  Run tox partially by argument.
        Need to specify the module names and target function to run tox.
* `-o`  Specify the output directory for the log files by argument.
* `-k`  Stop the tox process.
* `-v`  Show the version.
* `-h`  Show the help message.

### Example
* run all modules.
  ```
  autox.sh all
  ```
* run weko modules without weko-admin.
  ```
  autox.sh weko -n weko-admin
  ```
* Specify directory to export log.
  ```
  autox.sh -o example all
  ```
  ✔️ Test logs are output to log/ by default. Optionally, output can be specified to any directory under log/.
* run tox partially.
  ```
  autox.sh -p weko-admin test_api.py::test_is_restricted_user test_tasks.py::test_send_all_reports
  ```
  ✔️ Immediately following the -p option is treated as an optional argument, and everything after that is treated as a script argument.

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
### ver.1.2.0
add options: `-p`; :clap: Tests can now be run on a per-function basis.  
Coverage reports are now output after tests.

### ver.1.1.1
delete options: `-i` and `-w`.  
add options: `-o`; specify the output directory for the log files. `-v`; show the version.

### ver.1.1.0
add commands: all, invenio, weko.

### ver.1.0
the first script.
