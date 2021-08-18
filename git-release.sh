#!/usr/bin/env bash
set -e

find_latest_semver() {
    pattern="^v([0-9]+\.[0-9]+\.[0-9]+)\$"
    versions=$(for tag in $(git tag); do
        [[ "$tag" =~ $pattern ]] && echo "${BASH_REMATCH[1]}"
    done)
    if [ -z "$versions" ]; then
        echo 0.0.0
    else
        echo "$versions" | tr '.' ' ' | sort -nr -k 1 -k 2 -k 3 | tr ' ' '.' | head -1
    fi
}

increment_ver() {
    if [[ $4 == 'major' ]]; then
        awk_ptr='{printf("%d.%d.%d", $1+a, 0 , 0)}'
    elif [[ $4 == 'minor' ]]; then
        awk_ptr='{printf("%d.%d.%d", $1+a, $2+b , 0)}'
    elif [[ $4 == 'patch' ]]; then
        awk_ptr='{printf("%d.%d.%d", $1+a, $2+b , $3+c)}'
    fi

    find_latest_semver | \
        awk -F. -v a="$1" -v b="$2" -v c="$3" \
        "$awk_ptr"
}

bump() {
    next_ver="v$(increment_ver "$1" "$2" "$3" "$4")"
    git commit --allow-empty -m "$next_ver"
    git tag -a "$next_ver" -m "$next_ver"
    echo "Tagged $next_ver"
}

usage() {
    echo "Usage: git-release.sh {major|minor|patch}"
    echo "Bumps the semantic version field by one for a git-project."
    exit 1
}

while getopts h opt; do
    case $opt in
        h) usage;;
        :) echo "option -$OPTARG requires an argument"; exit 1;;
    esac
done

shift $((OPTIND-1))

case $1 in
    major) bump 1 0 0 major;;
    minor) bump 0 1 0 minor;;
    patch) bump 0 0 1 patch;;
    *) usage;;
esac
