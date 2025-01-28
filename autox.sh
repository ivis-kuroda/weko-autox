#!/bin/bash

# ==============================================================================
# autox.sh, ver.1.2.0
# update 2024.10.15
# Tomohiro KURODA
#
# Description:
#   This script is unofficial tool for running the unit tests of the Weko3 modules.
#
# Usage:
#   autox.sh [-n] [-r] [-p module] [-o output] [-k] [-v] [-h] [all|weko|invenio] [target1 target2 ...]
#
# Please refer to the README.md for more information.
# ==============================================================================

# current directory
CURRENT_DR=$(pwd)

# List of all modules
modules=()
for module in $CURRENT_DR/modules/*; do
    if [ -d "${module}/tests" ]; then
        modules+=($(basename $module))
    fi
done
# List of modules to be tested separately for each file
separately=("invenio-records" "weko-admin" "weko-authors" "weko-deposit" \
            "weko-items-ui" "weko-search-ui" "weko-records" "weko-records-ui" "weko-workflow")

reserved=("all" "invenio" "weko" "-i" "-w" "-n" "-r" "-k" "-o" "-v" "-h")

# kill the tox process
function cleanup() {
    echo 'Stopping tox process...'
    TOX_PIDS=$(docker-compose top web | grep 'tox' | awk '{print $2}')
    for PID in $TOX_PIDS; do
        kill $PID
        wait $PID 2>/dev/null
    done
    echo 'Tox process has been stopped.'
    cd $CURRENT_DR
    return 0
}
trap cleanup INT

function main(){

    targets=()
    n_flag=false
    r_flag=false
    p_flag=false
    output=""
    OPTIND=1
    # Parse the options
    while getopts "p:nkriwvho:" opt; do
        case $opt in
            n)
                n_flag=true
                ;;
            k)
                cleanup
                return 0
                ;;
            r)
                r_flag=true
                ;;
            p)
                p_flag=true
                targets=($OPTARG)
                ;;
            o)
                output=$OPTARG
                ;;
            v)
                echo "autox.sh - ver.1.2.0"
                return 0
                ;;
            h)
                echo "Usage:  autox.sh [-n] [-r] [-p module] [-o output] [-k] [-v] [-h] [all|weko|invenio] [target1 target2 ...]"
                echo ""
                echo "Commands:"
                echo "   all:     Run tests for all modules."
                echo "   invenio: Run tests for all invenio modules."
                echo "   weko:    Run tests for all weko modules."
                echo ""
                echo "Arguments:"
                echo "   target1 target2 ...  Specify the module names to run tox optionally."
                echo ""
                echo "Options:"
                echo "   -n  specify the module names to do not run tox by arguments."
                echo "   -r  Remove the egg-info and tox site-packages."
                echo "       When need to re-install the packages, or permission problems occur, use this option."
                echo "   -p  Run tox partially by argument."
                echo "       Need to specify the module names and target function to run tox."
                echo "   -o  Specify the output directory for the log files by argument."
                echo "   -k  Stop the tox process."
                echo "   -v  Show the version."
                echo "   -h  Show the help."
                echo ""
                echo "Please refer to the README.md for more information."
                return 0
                ;;
            \?)
                echo "Invalid option: -$OPTARG"
                return 1
                ;;
        esac
    done
    shift $((OPTIND -1))

    if [ $p_flag = false ]; then
        num_args="$#"
        if [[ $i_flag = true || " ${@} " =~ " invenio " ]]; then
            # Option -i or invenio is specified, add invenio modules to the targets
            for module in "${modules[@]}"; do
                if [[ "$module" =~ invenio ]]; then
                    targets+=("$module")
                fi
            done
        fi
        if [[ $w_flag = true  || " ${@} " =~ " weko " ]]; then
            # Option -w or weko is specified, add weko modules to the targets
            for module in "${modules[@]}"; do
                if [[ "$module" =~ weko ]]; then
                    targets+=("$module")
                fi
            done
        fi

        if [[ $n_flag = true || " ${@} " =~ " -n " ]]; then
            # Option -n is specified
            if [ ${#targets[@]} == 0 ]; then
                # No arguments provided.
                targets=("${modules[@]}")
            fi

            for arg in "$@"; do
                if [[ " ${modules[@]} " =~ " $arg " ]]; then
                    if [[ ! " ${targets[@]} " =~ " $arg " ]]; then
                        # do nothing if the module is not in the targets
                        continue
                    fi

                    # remove the module if it is in arguments
                    targets=("${targets[@]/$arg/}")
                else
                    if [[ " ${reserved[@]} " =~ " $arg " ]]; then
                        continue
                    fi

                    echo "unrecognized argument: '$arg'."
                    return 1
                fi
            done
            # remove empty elements
            targets=($(echo "${targets[@]}" | tr ' ' '\n' | grep -v '^$'))
        else
            if [[ " ${@} " =~ " all " ]]; then
                # if all is specified, add all modules to the targets
                targets=("${modules[@]}")
            else
                for arg in "$@"; do
                    # check if the name matches
                    if [[ " ${targets[@]} " =~ " $arg " ]]; then
                        # do nothing if the module is already in the targets
                        continue
                    fi

                    if [[ " ${modules[@]} " =~ " $arg " ]]; then
                        # add the module if it is in the modules
                        targets+=("$arg")
                    else
                        if [[ ! " ${reserved[@]} " =~ " $arg " ]]; then
                            echo "unrecognized argument: $arg."
                            return 1
                        fi
                    fi
                done
            fi
        fi
    fi

    OUTPUT_DR=""
    OBSERVE_DR=""
    if [ ${#targets[@]} == 0 ]; then
        echo "No modules specified."
        echo "If you want to see usage, please run 'autox.sh -h'."
        return 1
    elif [ ${#@} -eq 0 ]; then
        echo "invalid specifier."
        echo "If you want to see usage, please run 'autox.sh -h'."
        return 1
    else
        echo "${#targets[@]} module(s) found."
        if [ $output ]; then
            OUTPUT_DR="/code/log/$output"
            OBSERVE_DR="$CURRENT_DR/log/$output"
            mkdir -p $OBSERVE_DR
            echo "Output Directory: $CURRENT_DR/log/$output"
        else
            OUTPUT_DR="/code/log"
            OBSERVE_DR="$CURRENT_DR/log"
            echo "Output Directory: $CURRENT_DR/log"
        fi
    fi

    # install tox and tox-setuptools-version
    docker-compose exec web sh -c 'pip3 install tox==3.28; pip3 install tox-setuptools-version' > /dev/null 2>&1

    # loop through all targets subdirectories in the modules directory
    i=1
    for module in "${targets[@]}"; do
        printf "\r%$( tput cols )s\rSetup for $module. [$((i))/${#targets[@]} module(s)]"
        if [ $p_flag = false ]; then
            rm -rf $OBSERVE_DR/$module
            chown -R 1000:1000 $OBSERVE_DR
        fi
        if [ ! -d "$OBSERVE_DR/$module" ]; then
            mkdir -p $OBSERVE_DR/$module
        fi
        # erase the coverage data
        docker-compose exec web sh -c "cd /code/modules/$module; .tox/c1/bin/coverage erase" > /dev/null 2>&1
        cd $CURRENT_DR/modules/$module
        if [ $r_flag = true ]; then
            rm -rf $module.egg-info .tox htmlcov
            rm -f coverage.xml
        fi
        sleep 1

        if [ $p_flag = true ]; then
            # Run tox partially
            j=1
            for func in "$@"; do
                printf "\r%$( tput cols )s\r$func progressing. [$((j))/${#} function(s)]"
                docker-compose exec web sh -c "cd /code/modules/$module; .tox/c1/bin/pytest --cov=${module//-/\_} tests/$func -v -vv -s --cov-append --cov-branch --cov-report=term --cov-report=html --basetemp=/code/modules/$module/.tox/c1/tmp --full-trace > $OUTPUT_DR/$module/partial$j.log 2>&1" 2>/dev/null
                printf "\r%$( tput cols )s\r$func finished. [$((j))/${#} function(s)]\n"
                ((j++))
            done
        elif [[ " ${separately[@]} " =~ " $module " ]]; then
            # Run tox in the background to install the packages
            docker-compose exec -d web sh -c "cd /code/modules/$module; tox > $OUTPUT_DR/$module/install.log 2>&1" 2> /dev/null & disown
            printf "\r%$( tput cols )s\rInstalling packeges for the $module. [$((i))/${#targets[@]} module(s)]"
            sleep 10
            TOX_PIDS=$(docker-compose top web 2>/dev/null | grep 'tox' | awk '{print $2}')
            # Run tests separately for each file
            while [ -n "$TOX_PIDS" ]; do
                # Check if the installation is completed
                if grep -q '===='  $OBSERVE_DR/$module/install.log; then
                    printf "\r%$( tput cols )s\rInstalling packeges is completed."
                    for PID in $TOX_PIDS; do
                        kill $PID 2>/dev/null
                        wait $PID 2>/dev/null
                    done
                    rm -f $OBSERVE_DR/$module/install.log
                    break
                fi
                sleep 10
            done

            printf "\r%$( tput cols )s\r$module progressing. [$((i))/${#targets[@]} module(s)]\n"
            # get the list of test files
            mapfile -t files < <(find "${CURRENT_DR}/modules/${module}/tests" -name "test_*.py")
            j=1
            for file in "${files[@]}"; do
                # Run the test each file.
                file_name=$(basename $file)
                file_name=${file_name%.py}
                printf "\r%$( tput cols )s\r    $file_name.py [$((j))/${#files[@]} files]"
                docker-compose exec web sh -c "cd /code/modules/$module; .tox/c1/bin/pytest --cov=${module//-/\_} tests/$file_name.py -v --cov-append --cov-branch --cov-report=term --cov-report=html --basetemp=/code/modules/$module/.tox/c1/tmp --full-trace > $OUTPUT_DR/$module/$file_name.log 2>&1" 2>/dev/null
                ((j++))
            done
            printf "\r\e[1A"
        else
            # Run all tests together.
            printf "\r%$( tput cols )s\r$module progressing. [$((i))/${#targets[@]} module(s)]"
            docker-compose exec web sh -c "cd /code/modules/$module; tox > $OUTPUT_DR/$module/test_all.log 2>&1" 2>/dev/null
        fi

        coverage=""
        docker-compose exec web sh -c "cd /code/modules/$module; .tox/c1/bin/coverage report" > $OBSERVE_DR/$module/coverage.log 2> /dev/null
        coverage=$(cat $OBSERVE_DR/$module/coverage.log | grep TOTAL | awk '{print $NF}')
        printf "\r%$( tput cols )s\r$module finished. cov: \e[32m$coverage%\e[m [$((i))/${#targets[@]} module(s)]\n"
        ((i++))
    done

    echo 'All tests have been completed.'

}

main "$@"
cd $CURRENT_DR
