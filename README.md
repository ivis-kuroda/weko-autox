# weko-autox
weko auto unit test tool (unofficial)

## Description
This script is used to run the unit tests by tox of the Weko3 modules.

## Usage
`autox.sh [-n] [-k] [-r] [-i] [-w] [all|weko|invenio] [module1 module2 ...]`

### Commands
* `all`:     Run tests for all modules.
* `invenio`: Run tests for all invenio modules.
* `weko`:    Run tests for all weko modules.

### Arguments
`module1 module2 ...` : Specify the module names to run tox.

### Options
* `-i`  Run tests only the invenio modules.
* `-w`  Run tests only the weko modules.
* `-n`  specify the module names to do not run tox by arguments.
        Need to specify the module names to run tox.
* `-r`  Remove the egg-info and .tox directories.
        When permission problems occur, use this option.
* `-k`  Stop the tox process.
* `-h`  Show the help message.

## Note
* The log files are stored in the log directory.
* The following conditions must be satisfied in order for the progress o be displayed correctly
    - docker does not issue a warning.
    - The display must fit on a single line.

## Change Log
### ver.1.1
add commands: all, invenio, weko.

### ver.1.0
the first script.
