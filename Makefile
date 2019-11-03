#!/bin/bash

gen-test :
	mkdir example
	touch example/a.txt
	touch example/a.jpg
	touch example/a.pdf

rem-test :
	rm -rf example

reset-test: rem-test gen-test