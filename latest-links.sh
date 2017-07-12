#!/bin/bash


if [ $(git describe --tags | cut -c 1-1) == "1" ] ; then
	git describe --tags > release/1.x-latest
fi

if [ $(git describe --tags | cut -c 1-1) == "2" ] ; then
	git describe --tags > release/2.x-latest
fi
