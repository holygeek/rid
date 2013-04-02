#!/bin/sh
# Created: Mon Mar 11 13:23:26 MYT 2013

me=`basename $0`
usage() {
cat <<EOF
NAME
  $me - Show repository id (currently only grok git)

SYNOPSIS
  $me [-hr]

DESCRIPTION
  Show a repository unique id

OPTIONS
  -h
    Show this help message

  -r
    Align output to the right
EOF
}

align=left
while getopts hr opt
do
  case "$opt" in
    h) usage ; exit;;
    r) align=right;;
    \?) echo Unknown option ; exit;;
  esac
done
shift $(($OPTIND -1))

if test -d .git; then
  dirs=.
else
  dirs=`ls -d */|sort`
fi
reposig=`for d in $dirs; do
  GIT_DIR=$d/.git GIT_WORK_TREE=$d git log --no-decorate -1 --oneline
done | sort`

rightalign() {
  while read line; do
    for i in $(seq $((COLUMNS - ${#line}))); do
      echo -n " "
    done
    echo "$line"
  done
}

reposha1sum=`echo $reposig|sha1sum`

basename=`basename $PWD`
echo basename is $basename
if [ "$align" = "right" ]; then
  if [ -z "$COLUMNS" ]; then
    echo "$me: COLUMNS not set?"
    exit 1
  fi
  echo $basename $reposha1sum|rightalign
  echo $reposig|randomart|rightalign
else
  echo $basename $reposha1sum
  echo $reposig|randomart
fi
