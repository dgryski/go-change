#!/bin/sh

for w in 3 7 15 31
do
	for c in 0.1 0.3 0.6 0.8
	do
		echo "Marker width: ${w}; Correlation: ${c}"
		go run main.go -w=$w -c=$c -f=$1 > report-$w-$c.html &
	done

    wait
done

