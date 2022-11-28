#!/usr/bin/env sh

mkdir -p /tmp/testdir/{text,logs}
touch /tmp/testdir/file1.txt
touch /tmp/testdir/text/{text1,text2,text3}.txt
touch /tmp/testdir/logs/{log1,log2,log3}.log
tree /tmp/testdir
