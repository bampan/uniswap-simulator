#!/bin/bash

go build main.go
./main -n=2 -file=2_hours.json
./main -n=6 -file=6_hours.json
./main -n=24 -file=1_day.json
./main -n=168 -file=7_day.json
./main -n=720 -file=30_day.json
./main -n=20000000000 -file=non_compounding.json