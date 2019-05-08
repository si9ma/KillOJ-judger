#!/usr/bin/env bash

[ "$1" != "judge" ] && $@ && exit 0

mount -oremount,rw /sys/fs/cgroup
./kjudger ${@:2} # skip first arg