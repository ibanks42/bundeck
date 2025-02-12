#!/bin/bash

pkill -9 -f bundeck-dbg
cd web
bun run build
cd ..
go build -o ./bundeck-dbg .
