#!/usr/bin/env bash

[ "$1" != "judger" ] && $@ && exit 0

mount -oremount,rw /sys/fs/cgroup
./kjudger $@